package config

import (
	"crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"io"
	"strconv"
	"time"

	"github.com/joomcode/errorx"
	"github.com/yookoala/realpath"

	"github.com/project-illium/go-libp2p-tor-transport/internal/confStore"
)

// Check that all configurator are correctly done :
var _ = []Configurator{
	AllowTcpDial,
	DoSlowStart,
}

type Configurator func(*confStore.Config) error

// ConfMerge Merges different configs, starting at the first ending at the last.
func Merge(cs ...Configurator) Configurator {
	return func(c *confStore.Config) error {
		for _, v := range cs {
			if err := v(c); err != nil {
				return err
			}
		}
		return nil
	}
}

// AllowTcpDial allows the tor transport to dial tcp address.
// By Default TcpDial is off.
func AllowTcpDial(c *confStore.Config) error {
	c.AllowTcpDial = true
	return nil
}

// DoSlowStart set the tor node to bootstrap only when a Dial or a Listen is issued.
// By Default DoSlowStart is off.
func DoSlowStart(c *confStore.Config) error {
	c.TorStart.EnableNetwork = false
	return nil
}

// WithResourceManager sets a global resource manager for the transport
// otherwise it will be set to a null manager.
func WithResourceManager(rcmgr network.ResourceManager) Configurator {
	return func(c *confStore.Config) error {
		c.ResourceManager = rcmgr
		return nil
	}
}

// WithDNSPort optionally sets the tor DNS port setting
// to use DNS over tor.
func WithDNSPort(port int) Configurator {
	return func(c *confStore.Config) error {
		portStr := strconv.Itoa(port)
		c.TorStart.ExtraArgs = []string{"--DNSPort", portStr}
		return nil
	}
}

// WithPrivateKey provides an option to set the private key
// used when creating onion addresses.
//
// The key is an ed25519 key which you can create with crypto/ed25519.
//
// If this option is omitted the key will be random.
func WithPrivateKey(key crypto.PrivateKey) Configurator {
	return func(c *confStore.Config) error {
		c.PrivateKey = key
		return nil
	}
}

// SetSetupTimeout change the timeout for the bootstrap of the node and the publication of the tunnel.
// By Default SetupTimeout is at 5 minutes.
func SetSetupTimeout(t time.Duration) Configurator {
	return func(c *confStore.Config) error {
		if t == 0 {
			return errorx.IllegalArgument.New("Timeout can't be 0.")
		}
		c.SetupTimeout = t
		return nil
	}
}

// SetNodeDebug set the writer for the tor node debug output.
func SetNodeDebug(debug io.Writer) Configurator {
	return func(c *confStore.Config) error {
		c.TorStart.DebugWriter = debug
		return nil
	}
}

// SetBinaryPath set the path to the Tor's binary if you don't use the embeded Tor node.
func SetBinaryPath(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			return errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.ExePath = rpath
		return nil
	}
}

// SetDataDir sets the data directory where Tor is gonna put his
// data dir.
//
// If this isn't set a temp directory will be used.
func SetDataDir(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.DataDir = rpath
		return nil
	}
}

// SetTorrc sets the torrc file for tor to use instead of an blank one.
func SetTorrcPath(path string) Configurator {
	rpath, err := realpath.Realpath(path)
	return func(c *confStore.Config) error {
		if err != nil {
			errorx.Decorate(err, "Can't resolve path")
		}
		c.TorStart.TorrcFile = rpath
		return nil
	}
}
