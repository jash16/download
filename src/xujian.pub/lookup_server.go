package main

import (
    "fmt"
    "flag"
    "time"
    "os"
    "os/signal"
    "syscall"
    "xujian.pub/lookup"
    "xujian.pub/common/version"
)

var (
    ver = flag.Bool("version", false, "print version string")
//    verbose = flag.Bool("verbose", false, "enable verbose logging")
    tcpaddr = flag.String("tcp-address", "0.0.0.0:6790", "<addr>:<port> tcp listen address")
    httpaddr = flag.String("http-address", "0.0.0.0:6791", "<addr>:<port> http listen address")
    interTimeout = flag.Duration("interactive-producer-timeout", 300 *time.Second, "producer interactive timeout since last ping")
)

func main() {
    flag.Parse()
    opts := lookup.NewLookupOptions()

    if *ver {
        fmt.Println(version.LookupVersion)
        return
    }

    if tcpaddr != nil {
        opts.TcpAddress = *tcpaddr
    }

    if httpaddr != nil {
        opts.HttpAddress = *httpaddr
    }

    if interTimeout != nil {
        opts.InteractiveTimeout = *interTimeout
    }

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    ls := lookup.NewLookupServer(opts)
    ls.Main()

    <- sigChan

    ls.Exit()
}
