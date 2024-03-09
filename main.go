package main

import (
	"fmt"
	"go-chattery/server"
	"log/slog"
	"net"
)

const (
	ServerNetworkType = "tcp"
	ServerNetworkPort = 1337
)

func main() {
	l, err := net.Listen(ServerNetworkType, fmt.Sprintf(":%d", ServerNetworkPort))
	failure(err)
	defer l.Close()
	slog.Info("the server is now running", "address", l.Addr())

	chatServer := server.NewChatServer()

	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}

		chatServer.Join(conn)
	}
}

func failure(err error) {
	if err == nil {
		return
	}

	panic(err)
}
