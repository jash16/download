package main

import (
    "xujian.pub/common/version"
    "xujian.pub/server"
    "fmt"
    "flag"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func serverFlagSet() *flag.FlagSet {
    flagSet := flag.NewFlagSet("server", flag.ExitOnError)

    flagSet.Bool("version", false, "print version string")
    flagSet.Bool("verbose", false, "enable verbose logging")
    flagSet.String("tcp-address", "0.0.0.0:6789", "<addr>:<port> to listen")

    flagSet.String("data-path", "/data/download", "path where the data store")

    flagSet.Duration("client-timeout", 60 * time.Second, "client timeout")
    flagSet.Int64("max-clients", 100000, "how many clients cloud connect to this server")

    flagSet.Bool("cached", false, "cache some hot data")
    flagSet.Int64("cache-size", 1024*1024, "cache data per slice")
    flagSet.String("log-file", "", "the file to save the log, if null, output to the stderr")
    return flagSet
}


var (
    ver = flag.Bool("version", false, "print version string")
    verbose = flag.Bool("verbose", false, "enable verbose logging")
    tcpaddr = flag.String("tcp-address", "0.0.0.0:6789", "<addr>:<port> to listen")

    datapath = flag.String("data-path", "/data/download", "path where the data store")

    client_timeout = flag.Duration("client-timeout", 60 * time.Second, "client timeout")
    max_client = flag.Int64("max-clients", 100000, "how many clients cloud connect to this server")

    cached = flag.Bool("cached", false, "cache some hot data")
    cache_size = flag.Int64("cache-size", 1024*1024, "cache data per slice")
    log_file = flag.String("log-file", "", "the file to save the log, if null, output to the stderr")
)

func main() {
    flag.Parse()
    opt := server.NewServerOption()
    if *ver {
        fmt.Printf("download server v%s\n", version.ServerVersion)
        return
    }
    if *verbose {
        opt.Verbose = *verbose
    }
    if tcpaddr != nil {
        opt.TCPAddress = *tcpaddr
    }
    if datapath != nil {
        opt.DataPath = *datapath
    }
    if client_timeout != nil {
        opt.ClientTimeout = *client_timeout
    }
    if max_client != nil {
        opt.MaxClients = *max_client
    }

    if *cached {
        opt.Cached = *cached
    }
    if cache_size != nil {
        opt.CachedSize = *cache_size
    }

    if log_file != nil {
        opt.LogFile = *log_file
    }
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

    s := server.NewServer(opt)
    s.Main()
    <- signalChan
    s.Exit()
}
