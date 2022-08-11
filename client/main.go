package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var address = "localhost:2222"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	for {
		select {
		case <-done:
			return
		case t := <-time.After(time.Second):
			// connect to the relay server
			conn, err := net.Dial("tcp", address)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer conn.Close()

			msg := fmt.Sprintf("Tick %s!", t.String())
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Fatalln("write:", err.Error())
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatalln("read:", err.Error())
			}
			fmt.Printf("Received %d bytes from %s:\n%s\n", n, conn.RemoteAddr(), msg[:n])
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.

			// connect to the relay server
			conn, err := net.Dial("tcp", address)
			if err != nil {
				log.Println("dial:", err.Error())
				return
			}
			defer conn.Close()

			_, err = conn.Write([]byte("closing"))
			if err != nil {
				log.Println("write close:", err)
				return
			}

			var msg = make([]byte, 1024)
			var n int
			if n, err = conn.Read(msg); err != nil {
				log.Fatalln("read:", err.Error())
			}
			fmt.Printf("Received %d bytes from %s:\n%s\n", n, conn.RemoteAddr(), msg[:n])
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
