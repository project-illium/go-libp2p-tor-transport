package confStore

import (
	"github.com/libp2p/go-libp2p/core/network"
	"time"

	"github.com/cretz/bine/tor"
)

// `Config` stores the config, don't use it, you must use Configurator.
type Config struct {
	AllowTcpDial    bool
	SetupTimeout    time.Duration
	ResourceManager network.ResourceManager

	TorStart *tor.StartConf
}
