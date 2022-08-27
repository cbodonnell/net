package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/cbodonnell/net/pkg/net"
	tcpserver "github.com/cbodonnell/net/pkg/tcp/server"
	udpserver "github.com/cbodonnell/net/pkg/udp/server"
)

type ServerOpts struct {
	RelayAddress  string
	ServerAddress string
	ServerName    string
	Key           []byte
	RetryDuration string
	Debug         bool
}

func ServerCmd() error {
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

	var retryDuration string
	var debug bool

	serverCmd := flag.NewFlagSet(network, flag.ExitOnError)
	serverCmd.StringVar(&retryDuration, "retry-duration", "1s", "The duration to wait between retries")
	serverCmd.BoolVar(&debug, "debug", false, "Print debug messages")
	serverCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s %s %s [flags] <relayAddress> <serverName> <serverAddress>\n", os.Args[0], os.Args[1], os.Args[2])
		serverCmd.PrintDefaults()
	}
	serverCmd.Parse(os.Args[3:])

	relayAddress := serverCmd.Arg(0)
	if relayAddress == "" {
		return fmt.Errorf("relayAddress is required")
	}

	serverName := serverCmd.Arg(1)
	if serverName == "" {
		return fmt.Errorf("serverName is required")
	}

	serverAddress := serverCmd.Arg(2)
	if serverAddress == "" {
		return fmt.Errorf("serverAddress is required")
	}

	server, err := NewServer(network, ServerOpts{
		RelayAddress:  relayAddress,
		ServerAddress: serverAddress,
		ServerName:    serverName,
		Key:           []byte("passphrasewhichneedstobe32bytes!"),
		RetryDuration: retryDuration,
		Debug:         debug,
	})
	if err != nil {
		return fmt.Errorf("error creating server: %s", err.Error())
	}

	return server.Run()
}

func NewServer(network string, opts ServerOpts) (net.Server, error) {
	switch network {
	case "tcp":
		return tcpserver.NewTCPServer(tcpserver.TCPServerOpts{
			RelayAddress:  opts.RelayAddress,
			ServerAddress: opts.ServerAddress,
			// ServerName:    opts.ServerName,
			Key:           opts.Key,
			RetryDuration: opts.RetryDuration,
			Debug:         opts.Debug,
		})
	case "udp":
		return udpserver.NewUDPServer(udpserver.UDPServerOpts{
			RelayAddress:  opts.RelayAddress,
			ServerAddress: opts.ServerAddress,
			ServerName:    opts.ServerName,
			Key:           opts.Key,
			RetryDuration: opts.RetryDuration,
			Debug:         opts.Debug,
		})
	default:
		return nil, fmt.Errorf("unknown network: %s", network)
	}
}
