package ws

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/proxy"
	"golang.org/x/net/websocket"
)

type Dialer struct {
	Header http.Header
}

func (d *Dialer) Dial(ctx context.Context, url string) (*Conn, error) {
	return newConn(url, d.Header)
}

type Conn struct {
	conn *websocket.Conn

	status string
}

var _ proxy.Dialer = &proxyDialer{}

type proxyDialer struct {
	conn net.Conn
}

func (d *proxyDialer) Dial(network, addr string) (net.Conn, error) {
	return d.conn, nil
}

func newConn(address string, header http.Header) (*Conn, error) {
	config, err := websocket.NewConfig(address, "http://localhost")
	if err != nil {
		return nil, err
	}
	config.Header = header
	requestURI, err := url.ParseRequestURI(address)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", requestURI.Host)
	if err != nil {
		return nil, err
	}
	conn, err = proxy.FromEnvironmentUsing(&proxyDialer{conn: conn}).Dial("tcp", requestURI.Host)
	if err != nil {
		return nil, err
	}
	wsconn, err := websocket.NewClient(config, conn)
	if err != nil {
		return nil, err
	}

	return &Conn{conn: wsconn}, nil
}

func (c *Conn) Send(message string) error {
	return websocket.Message.Send(c.conn, message)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Conn) Read() (string, error) {
	var message string
	err := websocket.Message.Receive(c.conn, &message)
	if err != nil {
		c.status = "dead"
		return "", err
	}
	c.status = "alive"
	return message, nil
}

func (c *Conn) IsAlive() bool {
	return c.status == "alive"
}

func (c *Conn) Close() error {
	return c.conn.Close()
}
