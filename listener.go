package tor

import (
	"context"
	"github.com/libp2p/go-libp2p/core/network"
	"net"
	"strconv"
	"sync"

	"github.com/cretz/bine/tor"

	tpt "github.com/libp2p/go-libp2p/core/transport"

	ma "github.com/multiformats/go-multiaddr"

	"github.com/joomcode/errorx"
)

type listener struct {
	service *tor.OnionService
	ctx     context.Context
	cancel  func()
	closer  sync.Once

	upgrader tpt.Upgrader
	t        *transport
	rcmgr    network.ResourceManager

	lAddrStore *listenStore
	lAddrCur   *listenHolder
}

func (l *listener) Multiaddr() ma.Multiaddr {
	var base string
	if l.service.Version3 {
		base = "/onion3/"
	} else {
		base = "/onion/"
	}
	m, err := ma.NewMultiaddr(base + l.service.ID + ":" + strconv.Itoa(l.service.RemotePorts[0]))
	checkError(err)
	return m
}

func (l *listener) Addr() net.Addr {
	return addr(l.service.ID + ":" + strconv.Itoa(l.service.RemotePorts[0]))
}

func (l *listener) Close() error {
	var err error
	l.closer.Do(func() {
		// Remove the listener from the store.
		l.lAddrStore.Lock()
		cur := l.lAddrStore.cur
		if cur == l.lAddrCur {
			l.lAddrStore.cur = l.lAddrCur.next
			goto FinishRemovingLAddr
		}
		// No need to check for nil, we must hit our current first.
		for cur.next != l.lAddrCur {
			cur = cur.next
		}
		cur.next = l.lAddrCur.next
	FinishRemovingLAddr:
		l.lAddrStore.Unlock()
		l.cancel()
		err = l.service.Close()
	})
	return err
}

func (l *listener) Accept() (tpt.CapableConn, error) {
	c, err := l.service.Accept()
	if err != nil {
		return nil, errorx.Decorate(err, "Can't accept connection")
	}

	maconn := &listConn{
		netConnWithoutAddr: c,
		l:                  l,
		raddr:              NopMaddr2,
	}

	connScope, err := l.rcmgr.OpenConnection(network.DirInbound, true, maconn.RemoteMultiaddr())
	if err != nil {
		return nil, errorx.Decorate(err, "Resource manager blocked incoming tor connection")
	}

	conn, err := l.upgrader.Upgrade(
		l.ctx,
		l.t,
		maconn,
		network.DirInbound,
		"",
		connScope,
	)
	if err != nil {
		return nil, errorx.Decorate(err, "Can't upgrade raddr exchange connection")
	}

	stream, err := conn.AcceptStream()
	if err != nil {
		conn.Close()
		return nil, errorx.Decorate(err, "Can't accept raddr exchange conn")
	}

	// mbuf Maddr BUFfer
	var mbuf []byte
	buf := make([]byte, 40)
	var n, leftToRead int
	for {
		n, err = stream.Read(buf)
		if err != nil {
			// In this case terminate because there should be any reason this wouldn't
			// work.
			conn.Close()
			return nil, errorx.Decorate(err, "Can't read raddr exchange stream")
		}
		if n != 0 {
			break
		}
	}
	mbuf = buf[1:n]
	leftToRead -= n
	switch buf[0] {
	case encodeOnion:
		leftToRead += 39
	case encodeOnion3:
		leftToRead += 14
	default:
		// In case of default do nothing, it's not because we can't dial him back we
		// will not use this conn.
		goto EndLAddrExchange
	}
	for leftToRead > 0 {
		n, err = stream.Read(buf)
		if err != nil {
			// In this case terminate because there should be any reason this wouldn't
			// work.
			conn.Close()
			return nil, errorx.Decorate(err, "Can't read raddr exchange stream")
		}
		mbuf = append(mbuf, buf[:n]...)
		leftToRead -= n
	}
	{
		raddr, err := ma.NewMultiaddrBytes(mbuf)
		if err != nil {
			// In case of error do nothing, it's not because we can't dial him back we
			// will not use this conn.
			goto EndLAddrExchange
		}
		maconn.raddr = raddr
	}
EndLAddrExchange:
	stream.Close()
	return conn, nil
}
