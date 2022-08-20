package relay

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
)

type Relay struct {
	clientPort uint
	serverPort uint
	bufferSize uint
	debug      bool
}

type RelayOpts struct {
	ClientPort uint
	ServerPort uint
	BufferSize uint
	Debug      bool
}

func NewRelay(opts RelayOpts) *Relay {
	return &Relay{
		clientPort: opts.ClientPort,
		serverPort: opts.ServerPort,
		bufferSize: opts.BufferSize,
		debug:      opts.Debug,
	}
}

type Message struct {
	Data     []byte
	Response chan []byte
	Error    chan error
}

func (r Relay) Run() error {
	// Make a channel to handle errors.
	errChan := make(chan error)

	// Make a channel to handle messages.
	messageQueue := make(chan Message)

	clientPortString := fmt.Sprintf(":%d", r.clientPort)
	if r.debug {
		log.Printf("Listening for client on %s\n", clientPortString)
	}

	clientListener, err := net.Listen("tcp", clientPortString)
	if err != nil {
		return err
	}

	go r.handleClientConnections(clientListener, messageQueue, errChan)

	serverPortString := fmt.Sprintf(":%d", r.serverPort)
	if r.debug {
		log.Printf("Listening for server on %s\n", serverPortString)
	}

	serverListener, err := net.Listen("tcp", serverPortString)
	if err != nil {
		return err
	}

	go r.handleServerConnections(serverListener, messageQueue, errChan)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case <-interrupt:
		return errors.New("interrupted")
	case <-errChan:
		return fmt.Errorf("error: %s", <-errChan)
	}
}

func (r *Relay) handleClientConnections(clientListener net.Listener, messageQueue chan Message, errChan chan<- error) error {
	defer clientListener.Close()
	for {
		// Listen for an incoming connections from the client.
		conn, err := clientListener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("error accepting from client: %s", err.Error())
		}
		// Handle connections from the client.
		if err = r.handleClientRequest(conn, messageQueue); err != nil {
			fmt.Fprintf(os.Stderr, "error handling client request: %s", err.Error())
		}
	}
}

func (r *Relay) handleClientRequest(conn net.Conn, queue chan<- Message) error {
	// Close the connection when you're done with it.
	defer conn.Close()

	// Make a buffer to hold incoming data.
	if r.debug {
		log.Println("CLIENT: Reading from client")
	}
	buf := make([]byte, r.bufferSize)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading from client: %s", err.Error())
	}

	message := Message{
		Data:     buf[:reqLen],
		Response: make(chan []byte),
		Error:    make(chan error),
	}
	if r.debug {
		log.Printf("CLIENT: Received message with %d bytes:\n%s\n", reqLen, string(message.Data))
	}

	// add message to the queue
	if r.debug {
		log.Println("CLIENT: Adding message to queue")
	}
	queue <- message

	// wait for a response to the message
	if r.debug {
		log.Println("CLIENT: Waiting for a response")
	}

	var response []byte
	select {
	case response = <-message.Response:
	case err = <-message.Error:
		return fmt.Errorf("server returned an error: %s", err.Error())
	}

	if r.debug {
		log.Printf("CLIENT: Received response with %d bytes:\n%s\n", len(response), string(response))
	}

	// Send a response back to person contacting us.
	if r.debug {
		log.Println("CLIENT: Sending response to client")
	}
	_, err = conn.Write(response)
	if err != nil {
		return fmt.Errorf("error writing to client: %s", err.Error())
	}

	return nil
}

func (r *Relay) handleServerConnections(serverListener net.Listener, messageQueue chan Message, errChan chan<- error) error {
	defer serverListener.Close()
	for {
		// Listen for an incoming connections from the server.
		conn, err := serverListener.Accept()
		if err != nil {
			errChan <- fmt.Errorf("error accepting from server: %s", err.Error())
		}

		if r.debug {
			log.Println("SERVER: Reading from message queue")
		}
		message := <-messageQueue

		// Handle connections from the server.
		if err = r.handleServerRequest(conn, message); err != nil {
			// return fmt.Errorf("error handling server request: %s", err.Error())
			message.Error <- fmt.Errorf("error handling server request: %s", err.Error())
			fmt.Fprintf(os.Stderr, "error handling server request: %s", err.Error())
		}
	}
}

func (r *Relay) handleServerRequest(conn net.Conn, message Message) error {
	// Close the connection when you're done with it.
	defer conn.Close()

	if r.debug {
		log.Printf("SERVER: Found message with %d bytes:\n%s\n", len(message.Data), string(message.Data))
	}

	// Send the message data to the server component.
	if r.debug {
		log.Println("SERVER: Sending message to server")
	}
	_, err := conn.Write(message.Data)
	if err != nil {
		return fmt.Errorf("error writing to server: %s", err.Error())
	}

	// Make a buffer to hold incoming data.
	if r.debug {
		log.Println("SERVER: Reading from server")
	}
	buf := make([]byte, r.bufferSize)
	// Read the incoming connection into the buffer.
	reqLen, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("error reading from server: %s", err.Error())
	}

	response := buf[:reqLen]
	if r.debug {
		log.Printf("SERVER: Received response with %d bytes:\n%s\n", reqLen, string(response))
	}

	// data was received from the server, respond to the message
	if r.debug {
		log.Println("SERVER: Sending response to message")
	}
	message.Response <- response

	return nil
}
