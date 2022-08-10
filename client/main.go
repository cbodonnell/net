package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var address = "localhost:4223"

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	var d net.Dialer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		log.Fatalf("Failed to dial: %v", err)
	}
	defer conn.Close()

	done := make(chan struct{})
	go func(c net.Conn, done chan<- struct{}) {
		defer close(done)
		for {
			var msg = make([]byte, 1024)
			var n int
			if n, err = c.Read(msg); err != nil {
				log.Fatalln("read:", err.Error())
			}
			fmt.Printf("Received %d bytes from %s:\n%s\n", n, c.RemoteAddr(), msg[:n])
		}
	}(conn, done)

	for {
		select {
		case <-done:
			return
		case t := <-time.After(time.Second):
			msg := fmt.Sprintf("Tick %s!", t.String())
			if _, err := conn.Write([]byte(msg)); err != nil {
				log.Fatalln("write:", err.Error())
			}
		case <-interrupt:
			log.Println("interrupt")
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			_, err := conn.Write([]byte("closing"))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
