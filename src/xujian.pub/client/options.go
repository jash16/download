package client

import (
    "xujian.pub/common"
    "log"
    "os"
    "time"
)

type ClientOption struct {
    Version    bool
    Verbose    bool
    ServerAddr string `host:port`
    LookupAddr string
    Files  []string
    SaveDir string
    RoutingNum int64
    BlockSize  int64
    MaxWait    time.Duration
    Logger     common.Logger
}

func NewClientOption() *ClientOption {
    return &ClientOption{
        Logger: log.New(os.Stderr, "[MyClient] ", log.Ldate|log.Ltime|log.Lmicroseconds),
    }
}
