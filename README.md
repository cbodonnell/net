# E2E Encrypted Client-Server Message Queue

## Client - Listener & Dialer

Client-side component listens to the client and dials to the relay.

It listens for messages from the client application and dials them
to the relay. A response is returned from the relay.

## Relay - Listener & Listener

Relay listens to the client and server and relays the messages.

It listens for messages from the client and places them in a channel
for the server to receive. A response is returned to the client
after the server has taken the message and responded.

## Server - Dialer & Dialer

Server-side component dials to the relay and dials to the server.

It dials to the relay and receives a response that is then dialed
to the server application. It sends responses from the server
to the relay when dialing for the next message.