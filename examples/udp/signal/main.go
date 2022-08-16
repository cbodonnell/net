package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// TODO: Should this be a map of servers and their client list? Could relay responses to all clients.
var clients = make(map[string]string)

var servers = make(map[string]string)
var monitors = make(map[string]chan *net.UDPAddr)

func main() {
	localAddress := ":9595"
	if len(os.Args) > 2 {
		localAddress = os.Args[2]
	}

	addr, err := net.ResolveUDPAddr("udp", localAddress)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Printf("Listening on %s\n", conn.LocalAddr().String())

	for {
		buffer := make([]byte, 1024)
		bytesRead, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("[INCOMING]", string(buffer[0:bytesRead]))

		parts := strings.Split(string(buffer[0:bytesRead]), ": ")
		if len(parts) != 2 {
			if _, err = conn.WriteToUDP([]byte("FAIL: BAD REQUEST"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write BAD REQUEST response %s\n", err.Error())
			}
			continue
		}

		action, target := parts[0], parts[1]

		switch action {
		case "PUNCH":
			// can only punch to a registered server
			if _, ok := servers[target]; !ok {
				if _, err = conn.WriteToUDP([]byte("FAIL: NOT REGISTERED"), remoteAddr); err != nil {
					fmt.Printf("[ERROR] Failed to write NOT REGISTERED response %s\n", err.Error())
				}
				continue
			}
			fmt.Printf("[PUNCH] from %s to %s\n", remoteAddr.String(), target)
			clients[remoteAddr.String()] = target
			if _, err = conn.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", servers[target])), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write PUNCH response %s\n", err.Error())
			}
		case "CLOSE":
			fmt.Printf("[CLOSE] %s closed\n", remoteAddr.String())
			delete(clients, remoteAddr.String())
			if _, err = conn.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write CLOSE response %s\n", err.Error())
			}
		case "REGISTER":
			if _, ok := servers[target]; ok {
				if _, err = conn.WriteToUDP([]byte("FAIL: ALREADY REGISTERED"), remoteAddr); err != nil {
					fmt.Printf("[ERROR] Failed to write ALREADY REGISTERED response %s\n", err.Error())
				}
				continue
			}
			fmt.Printf("[REGISTER] %s registered as %s\n", remoteAddr.String(), target)
			servers[target] = remoteAddr.String()

			// start ping loop
			ping := make(chan *net.UDPAddr)
			monitors[target] = ping
			go monitor(conn, target, ping)

			if _, err = conn.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write REGISTER response %s\n", err.Error())
			}
		case "PING":
			if _, ok := servers[target]; !ok {
				if _, err = conn.WriteToUDP([]byte("FAIL: NOT REGISTERED"), remoteAddr); err != nil {
					fmt.Printf("[ERROR] Failed to write NOT REGISTERED response %s\n", err.Error())
				}
				continue
			}
			if _, ok := monitors[target]; !ok {
				if _, err = conn.WriteToUDP([]byte("FAIL: NOT MONITORING"), remoteAddr); err != nil {
					fmt.Printf("[ERROR] Failed to write NOT MONITORING response %s\n", err.Error())
				}
				continue
			}
			if servers[target] != remoteAddr.String() {
				fmt.Printf("[ERROR] %s is not %s\n", remoteAddr.String(), target)
				continue
			}
			monitors[target] <- remoteAddr
		case "UNREGISTER":
			fmt.Printf("[UNREGISTER] %s unregistered\n", target)
			delete(servers, target)
			if _, err = conn.WriteToUDP([]byte(fmt.Sprintf("SUCCESS: %s", target)), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write UNREGISTER response %s\n", err.Error())
			}
		default:
			if _, err = conn.WriteToUDP([]byte("FAIL: BAD REQUEST"), remoteAddr); err != nil {
				fmt.Printf("[ERROR] Failed to write BAD REQUEST response %s\n", err.Error())
			}
		}
	}
}

func monitor(conn *net.UDPConn, target string, ping chan *net.UDPAddr) {
	for {
		select {
		case <-time.After(time.Second * 10):
			delete(servers, target)
			fmt.Printf("[UNREGISTER] %s unregistered after timeout\n", target)
			return
		case <-ping:
			fmt.Printf("[PING] %s\n", target)
		}
	}
}
