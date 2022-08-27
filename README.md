# Networking Utilities

## Client - Listener & Dialer

Client-side component listens for messages from the client application
and dials them to the relay server. It sends responses from the 
relay server to the client application.

## Relay - Listener & Listener

Relay server listens for messages from the client and places them in a
channel for the server to receive. A response is returned to the client
after the server has taken the message and responded.

## Server - Dialer & Dialer

Server-side component dials to the relay and receives a response
that is then dialed to the server application. It sends responses
from the server to the relay when dialing for the next message.

## TODO:

In no particular order:
* Registration for TCP
* Fallback to relay for UDP if punchthrough fails
* Encryption for UDP
* Authentication and Authorization mechanism