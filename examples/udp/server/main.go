package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func random(min, max int) int {
	return rand.Intn(max-min) + min
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalln("Please provide a port number!")
		return
	}
	PORT := ":" + os.Args[1]

	s, err := net.ResolveUDPAddr("udp4", PORT)
	if err != nil {
		log.Fatal(err)
	}

	connection, err := net.ListenUDP("udp4", s)
	if err != nil {
		log.Fatal(err)
	}

	defer connection.Close()
	buffer := make([]byte, 1024)
	rand.Seed(time.Now().Unix())

	for {
		n, addr, err := connection.ReadFromUDP(buffer)
		fmt.Print("-> ", string(buffer[0:n-1]))

		if strings.TrimSpace(string(buffer[0:n])) == "STOP" {
			fmt.Println("Exiting UDP server!")
			return
		}

		data := []byte(strconv.Itoa(random(1, 1001)))
		fmt.Printf("data: %s\n", string(data))
		_, err = connection.WriteToUDP(data, addr)
		if err != nil {
			log.Fatal(err)
		}
	}
}
