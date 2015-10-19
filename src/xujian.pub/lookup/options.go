package server

import (
    "log"
    "time"
    "xujian.pub/common"
)

type LookupOptions struct {
    TcpAddress string
    HttpAddress string
    InteractiveTimeout time.Duration
    Logger common.Logger
}

func NewLookupOptions() *LookupOptions {
    return &LookupOptions {
        TcpAddress: ":6790",
        HttpAddress: ":6791",
        InteractiveTimeout: 300*time.Second,
        Logger: log.New(os.Stderr, "MyLookup", log.Ldate|log.Ltime|log.Lmicroseconds)
    }
}
