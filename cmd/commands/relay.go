package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/cbodonnell/net/pkg/net"
	tcprelay "github.com/cbodonnell/net/pkg/tcp/relay"
	udprelay "github.com/cbodonnell/net/pkg/udp/relay"
)

type RelayOpts struct {
	ClientPort uint
	ServerPort uint
	BufferSize uint
	Debug      bool
}

func RelayCmd() error {
	rootCmd := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	rootCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s <network[tcp|udp]>\n", os.Args[0], os.Args[1])
		rootCmd.PrintDefaults()
	}
	rootCmd.Parse(os.Args[2:])

	network := rootCmd.Arg(0)
	if network == "" {
		return fmt.Errorf("network is required")
	}

	var clientPort uint
	var serverPort uint
	var bufferSize uint
	var debug bool

	relayCmd := flag.NewFlagSet(network, flag.ExitOnError)
	relayCmd.UintVar(&clientPort, "client-port", 3333, "The port to listen for the client on")
	relayCmd.UintVar(&serverPort, "server-port", 4444, "The port to listen for the server on")
	relayCmd.UintVar(&bufferSize, "buffer-size", 1024, "The maximum size of the buffer")
	relayCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	relayCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s [flags]\n", os.Args[0], os.Args[1], os.Args[2])
		relayCmd.PrintDefaults()
	}
	relayCmd.Parse(os.Args[3:])

	relay, err := NewRelay(network, RelayOpts{
		ClientPort: clientPort,
		ServerPort: serverPort,
		BufferSize: bufferSize,
		Debug:      debug,
	})
	if err != nil {
		return fmt.Errorf("error creating relay: %s", err.Error())
	}

	return relay.Run()
}

func NewRelay(network string, opts RelayOpts) (net.Relay, error) {
	switch network {
	case "tcp":
		return tcprelay.NewTCPRelay(tcprelay.TCPRelayOpts{
			ClientPort: opts.ClientPort,
			ServerPort: opts.ServerPort,
			BufferSize: opts.BufferSize,
			Debug:      opts.Debug,
		}), nil
	case "udp":
		return udprelay.NewUDPRelay(udprelay.UDPRelayOpts{
			ClientPort: opts.ClientPort,
			ServerPort: opts.ServerPort,
			BufferSize: opts.BufferSize,
			Debug:      opts.Debug,
		}), nil
	default:
		return nil, fmt.Errorf("unknown network: %s", network)
	}
}
