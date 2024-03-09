package server

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
	id   string

	serverInChannel chan<- ClientInPayload
	outgoing        chan string
	reader          *bufio.Reader
	writer          *bufio.Writer
}

func NewClient(conn net.Conn) *Client {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	clientID := conn.RemoteAddr().String()

	c := &Client{
		outgoing: make(chan string),
		conn:     conn,
		id:       clientID,
		reader:   reader,
		writer:   writer,
	}

	go c.Write()
	return c
}

func (c *Client) WithInChannel(ch chan<- ClientInPayload) *Client {
	c.serverInChannel = ch

	go c.Read()
	return c
}

func (c *Client) SendMessage(msg string) {
	c.outgoing <- msg
}

func (c *Client) Write() {
	for outgoingMessage := range c.outgoing {
		_, err := c.writer.WriteString(outgoingMessage)
		if err != nil {
			slog.Error("something went wrong when writing outgoing message to client's writer",
				"clientID", c.id,
				"err", err,
				"outgoingMessage", outgoingMessage,
			)
			continue
		}

		err = c.writer.Flush()
		if err != nil {
			slog.Error("something went wrong when flushing the client's writer",
				"clientID", c.id,
				"err", err,
				"outgoingMessage", outgoingMessage,
			)
			continue
		}
	}
}

func (c *Client) Read() {
	for {
		bytes, err := c.reader.ReadBytes('\n')

		// generate new message payload
		if err == nil {
			isEmptyMessage := len(strings.TrimSuffix(string(bytes), "\n")) == 0
			if isEmptyMessage {
				continue
			}

			c.serverInChannel <- ClientInPayload{
				ClientID: c.id,
				Type:     NewMessage,
				Data:     bytes,
			}
			continue
		}

		// generate disconnect payload
		isDisconnectSignal := err == io.EOF
		if isDisconnectSignal {
			c.serverInChannel <- ClientInPayload{
				ClientID: c.id,
				Type:     Disconnect,
			}
			break
		}

		slog.Error("something went wrong when reading client incoming message",
			"clientID", c.id,
			"err", err,
			"bytesIn", bytes,
		)
	}
}

func (c *Client) Quit() error {
	return c.conn.Close()
}
