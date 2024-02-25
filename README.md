# go-libp2p-tor-transport
Go tor transport is a [go-libp2p](https://github.com/libp2p/go-libp2p) transport forked from berty.tech.
We've removed the embedded tor instance as it was outdated and unlikely to ever be current with
mainline tor.

### With config :
```go
import (
  "context"
  "time"

  tor "github.com/project-illium/go-libp2p-tor-transport"
  config "github.com/project-illium/go-libp2p-tor-transport/config"
  libp2p "github.com/libp2p/go-libp2p"
)

func main() {
  builder, err := tor.NewBuilder(        // NewBuilder can accept some `config.Configurator`
    config.AllowTcpDial,                 // Some Configurator are already ready to use.
    config.SetSetupTimeout(time.Minute), // Some require a parameter, in this case it's a function that will return a Configurator.
    config.SetBinaryPath("/usr/bin/tor"),
  )
  // Evrything else is as previously shown.
  c(err)
  hostWithConfig, err := libp2p.New(
    context.Background(),
    libp2p.Transport(builder),
  )
  c(err)
}

func c(err error) {
  if err != nil {
    panic(err)
  }
}
```