package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/cbodonnell/net/pkg/net"
	tcpclient "github.com/cbodonnell/net/pkg/tcp/client"
	udpclient "github.com/cbodonnell/net/pkg/udp/client"
)

type ClientOpts struct {
	Port         uint
	RelayAddress string
	ServerName   string
	Key          []byte
	BufferSize   uint
	Debug        bool
}

func ClientCmd() error {
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

	var port uint
	var bufferSize uint
	var debug bool

	clientCmd := flag.NewFlagSet(network, flag.ExitOnError)
	clientCmd.UintVar(&port, "port", 2222, "The port to listen on")
	clientCmd.UintVar(&bufferSize, "buffer-size", 1024, "The maximum size of the buffer")
	clientCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	clientCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s [flags] <relayAddress> <serverName>\n", os.Args[0], os.Args[1], os.Args[2])
		rootCmd.PrintDefaults()
		clientCmd.PrintDefaults()
	}
	clientCmd.Parse(os.Args[3:])

	relayAddress := clientCmd.Arg(0)
	if relayAddress == "" {
		return fmt.Errorf("relayAddress is required")
	}

	serverName := clientCmd.Arg(1)
	if relayAddress == "" {
		return fmt.Errorf("serverName is required")
	}

	client, err := NewClient(network, ClientOpts{
		Port:         port,
		RelayAddress: relayAddress,
		ServerName:   serverName,
		Key:          []byte("passphrasewhichneedstobe32bytes!"),
		BufferSize:   bufferSize,
		Debug:        debug,
	})
	if err != nil {
		return fmt.Errorf("error creating client: %s", err.Error())
	}

	return client.Run()
}

func NewClient(network string, opts ClientOpts) (net.Client, error) {
	switch network {
	case "tcp":
		return tcpclient.NewTCPClient(tcpclient.TCPClientOpts{
			Port:         opts.Port,
			RelayAddress: opts.RelayAddress,
			// ServerName:   opts.ServerName,
			Key:        opts.Key,
			BufferSize: opts.BufferSize,
			Debug:      opts.Debug,
		})
	case "udp":
		return udpclient.NewUDPClient(udpclient.UDPClientOpts{
			Port:         opts.Port,
			RelayAddress: opts.RelayAddress,
			ServerName:   opts.ServerName,
			Key:          opts.Key,
			BufferSize:   opts.BufferSize,
			Debug:        opts.Debug,
		})
	default:
		return nil, fmt.Errorf("unknown network: %s", network)
	}
}
