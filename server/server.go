package server

import (
	"log/slog"
	"net"
)

type ClientInPayloadType uint8

const (
	NewMessage ClientInPayloadType = iota
	Disconnect
)

type ClientInPayload struct {
	ClientID string
	Type     ClientInPayloadType
	Data     any
}

type ChatServer struct {
	clients map[string]*Client

	inChannel chan ClientInPayload
}

func NewChatServer() *ChatServer {
	cs := &ChatServer{
		clients:   make(map[string]*Client),
		inChannel: make(chan ClientInPayload),
	}

	go cs.ListenInPayload()
	return cs
}

func (cs *ChatServer) ListenInPayload() {
	for inPayload := range cs.inChannel {
		switch inPayload.Type {
		case NewMessage:
			message := string(inPayload.Data.([]byte))
			slog.Info("received a message from the client",
				"clientID", inPayload.ClientID,
				"message", message,
			)
		case Disconnect:
			cs.Disconnect(inPayload.ClientID)
		}
	}
}

func (cs *ChatServer) Disconnect(clientID string) {
	client, isExist := cs.clients[clientID]
	if !isExist {
		slog.Warn("could not disconnect a client because it does not exist or disconnected already",
			"clientID", client.id,
		)
		return
	}

	close(client.outgoing)
	err := client.Quit()
	if err != nil {
		slog.Error("an error occurred when quiting the client",
			"clientID", client.id,
			"err", err,
		)
	}
}

func (cs *ChatServer) Join(conn net.Conn) {
	client := NewClient(conn).WithInChannel(cs.inChannel)

	cs.clients[client.id] = client
	client.SendMessage("welcome to the server!")
}
