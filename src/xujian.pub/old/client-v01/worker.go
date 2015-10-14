package client

import (
    "net"
    "io"
)

type downConn struct {
    conn net.Conn
    r io.Reader
    w io.Reader
}

type worker struct {
    reqConn net.Conn
    r io.Reader
    w io.Writer
    file string
    client *Client
    conns []*downConn
}
