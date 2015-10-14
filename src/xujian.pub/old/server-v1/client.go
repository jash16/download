package server

import (
    "bufio"
    "sync"
    "time"
    "net"
)

const (
    defaultBufferSize int = 1048704 // 1048576 + 128
)

type Client struct {
    Addr string

    LastAccess int64
    AliveInternal int64

    Files map[string]*FileInfo
    conn net.Conn
    Reader *bufio.Reader
    Writer *bufio.Writer

    sync.RWMutex

    ctx *context
}

func NewClient(conn net.Conn, ctx *context) *Client {

    client := new(Client)
    addr := conn.RemoteAddr().String()

    client.Addr = addr
    client.Reader = bufio.NewReaderSize(conn, defaultBufferSize)
    client.Writer = bufio.NewWriterSize(conn, defaultBufferSize)
    client.conn = conn
    client.Files = make(map[string]*FileInfo)
    client.LastAccess = time.Now().Unix()
    client.ctx = ctx

    return client
}

func (c *Client) UpdateAccessTime() {
    c.Lock()
    c.LastAccess = time.Now().Unix()
    c.RUnlock()
}

func (c *Client) IsAlive() bool {
    c.RLock()
    defer c.RUnlock()
    if c.LastAccess < time.Now().Unix() - c.AliveInternal {
        return false
    }
    return true
}
