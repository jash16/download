package server

import (
    "os"
    "time"
    "log"
    "xujian.pub/common"
     _ "flag"
)

type ServerOption struct {
    //server info host:port
    TCPAddress string `flag:"tcp-address"`
    HttpAddress string `flag:"http-address"`
    Verbose bool       `flag:"verbose"`
    LookupSrvAddrs []string `flag:"lookup-tcp-address"`

    //cache
    Cached bool        `flag:"cached"`
    CachedSize int64   `flag:"cache-size"`

    //client
    MaxClients int64   `flag:"max-clients"`
    ClientTimeout time.Duration

    //the files path
    DataPath string     `flag:"data-path"`

    //Logger
    Logger common.Logger
    LogFile string
}

func NewServerOption() *ServerOption {
    return &ServerOption {
        TCPAddress: "0.0.0.0:6789",
        MaxClients: 100000,
        ClientTimeout: 60 * time.Second,
        Cached:     false,
        CachedSize: 0,
        DataPath:   "/data/download/",
        Logger:     log.New(os.Stderr, "[MyServer]", log.Ldate|log.Ltime|log.Lmicroseconds),
    }
}
