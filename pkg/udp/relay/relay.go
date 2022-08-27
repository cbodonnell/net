package relay

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type UDPRelay struct {
	clientPort uint
	serverPort uint
	bufferSize uint
	debug      bool
	// clients    map[string]string
	servers  map[string]string
	monitors map[string]chan *net.UDPAddr
}

type UDPRelayOpts struct {
	ClientPort uint
	ServerPort uint
	BufferSize uint
	Debug      bool
}

func NewUDPRelay(opts UDPRelayOpts) *UDPRelay {
	return &UDPRelay{
		clientPort: opts.ClientPort,
		serverPort: opts.ServerPort,
		bufferSize: opts.BufferSize,
		debug:      opts.Debug,
		// clients:    make(map[string]string),
		servers:  make(map[string]string),
		monitors: make(map[string]chan *net.UDPAddr),
	}
}

func (r *UDPRelay) Run() error {
	// TODO: Listen on separate ports for clients and servers.
	clientPortString := fmt.Sprintf(":%d", r.clientPort)

	clientAddr, err := net.ResolveUDPAddr("udp", clientPortString)
	if err != nil {
		return fmt.Errorf("error resolving client port: %s", err)
	}
	clientListener, err := net.ListenUDP("udp", clientAddr)
	if err != nil {
		return fmt.Errorf("error listening on client port: %s", err)
	}
	defer clientListener.Close()

	if r.debug {
		fmt.Printf("Listening on %s\n", clientListener.LocalAddr().String())
	}

	for {
		if err := r.handleRequest(clientListener); err != nil {
			return fmt.Errorf("error handling connection: %s", err)
		}
	}
}

func (r *UDPRelay) handleRequest(clientListener *net.UDPConn) error {
	buffer := make([]byte, 1024)
	bytesRead, remoteAddr, err := clientListener.ReadFromUDP(buffer)
	if err != nil {
		return fmt.Errorf("failed to read from UDP: %s", err.Error())
	}

	if r.debug {
		fmt.Println("[INCOMING]", string(buffer[0:bytesRead]))
	}

	parts := strings.Split(string(buffer[0:bytesRead]), ": ")
	if len(parts) != 2 {
		if _, err = clientListener.WriteToUDP([]byte("FAIL: BAD REQUEST"), remoteAddr); err != nil {
			fmt.Printf("[ERROR] Failed to write BAD REQUEST response %s\n", err.Error())
		}
		return fmt.Errorf("invalid request: %s", string(buffer[0:bytesRead]))
	}

	action, target := parts[0], parts[1]

	if err := r.handleAction(action, target, clientListener, remoteAddr); err != nil {
		fmt.Printf("[ERROR] error handling action: %s\n", err.Error())
	}

	return nil
}

func (r *UDPRelay) handleAction(action, target string, clientListener *net.UDPConn, remoteAddr *net.UDPAddr) error {
	switch action {
	case "PUNCH":
		// can only punch to a registered server
		if _, ok := r.servers[target]; !ok {
			if _, err := clientListener.WriteToUDP([]byte("FAIL: NOT REGISTERED"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write NOT REGISTERED response %s\n", err.Error())
			}
			return fmt.Errorf("target not registered: %s", target)
		}

		if r.debug {
			fmt.Printf("[PUNCH] from %s to %s\n", remoteAddr.String(), target)
		}

		// r.clients[remoteAddr.String()] = target

		if _, err := clientListener.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", r.servers[target])), remoteAddr); err != nil {
			fmt.Printf("[ERROR] Failed to write PUNCH response %s\n", err.Error())
		}
	// case "CLOSE":
	// 	if r.debug {
	// 		fmt.Printf("[CLOSE] %s closed\n", remoteAddr.String())
	// 	}

	// 	delete(r.clients, remoteAddr.String())

	// 	if _, err := clientListener.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
	// 		fmt.Printf("[ERROR] Failed to write CLOSE response %s\n", err.Error())
	// 	}
	case "REGISTER":
		if _, ok := r.servers[target]; ok {
			if _, err := clientListener.WriteToUDP([]byte("FAIL: ALREADY REGISTERED"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write ALREADY REGISTERED response %s\n", err.Error())
			}
			return fmt.Errorf("target already registered: %s", target)
		}

		if r.debug {
			fmt.Printf("[REGISTER] %s registered as %s\n", remoteAddr.String(), target)
		}

		r.servers[target] = remoteAddr.String()

		// start ping loop
		ping := make(chan *net.UDPAddr)
		r.monitors[target] = ping
		go r.monitor(clientListener, target, ping)

		if _, err := clientListener.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
			fmt.Printf("[ERROR] Failed to write REGISTER response %s\n", err.Error())
		}
	case "PING":
		if _, ok := r.servers[target]; !ok {
			if _, err := clientListener.WriteToUDP([]byte("FAIL: NOT REGISTERED"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write NOT REGISTERED response %s\n", err.Error())
			}
			return fmt.Errorf("target not registered: %s", target)
		}
		if _, ok := r.monitors[target]; !ok {
			if _, err := clientListener.WriteToUDP([]byte("FAIL: NOT MONITORING"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write NOT MONITORING response %s\n", err.Error())
			}
			return fmt.Errorf("not monitoring target: %s", target)
		}
		if r.servers[target] != remoteAddr.String() {
			return fmt.Errorf("%s is not %s", remoteAddr.String(), target)
		}
		r.monitors[target] <- remoteAddr
	case "UNREGISTER":
		if r.debug {
			fmt.Printf("[UNREGISTER] %s unregistered\n", target)
		}
		delete(r.servers, target)
		if _, err := clientListener.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
			fmt.Printf("[ERROR] Failed to write UNREGISTER response %s\n", err.Error())
		}
	default:
		if _, err := clientListener.WriteToUDP([]byte("FAIL: BAD REQUEST"), remoteAddr); err != nil {
			fmt.Printf("[ERROR] Failed to write BAD REQUEST response %s\n", err.Error())
		}
	}
	return nil
}

func (r *UDPRelay) monitor(conn *net.UDPConn, target string, ping chan *net.UDPAddr) {
	for {
		select {
		case <-time.After(time.Second * 10):
			delete(r.servers, target)
			if r.debug {
				fmt.Printf("[UNREGISTER] %s unregistered after timeout\n", target)
			}
			return
		case server := <-ping:
			if r.debug {
				fmt.Printf("[PING] %s\n", target)
			}
			conn.WriteTo([]byte(fmt.Sprintf("PONG: %s", target)), server)
		}
	}
}
