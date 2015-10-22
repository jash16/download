package main

import (
    "os"
    "os/signal"
    "syscall"
    "fmt"
    "time"
    "flag"
    "strings"
    "xujian.pub/client"
    _"xujian.pub/common/util"
    "xujian.pub/common/version"
)

var (
    ver = flag.Bool("version", false, "print version")
    verbose = flag.Bool("verbose", false, "enable verbose logging")
    serverAddr = flag.String("server-tcp-address", "", "server tcp address")
    lookupAddr = flag.String("lookup-tcp-address", "", "lookup server tcp address")
    file = flag.String("download-files", "", "download file name, use ';' seperate, a.txt;b.txt")
    maxWait = flag.Duration("max-wait-time", 60*time.Second, "max time wait server send data")
    routings = flag.Int64("routing-num", 10, "routings per file")
    blockSize = flag.Int64("block-size", 1048576, "size per block")
    saveDir = flag.String("save-dir", "./", "dir to save download file")
)

func main() {

    flag.Parse()
    opt := client.NewClientOption()
    if *ver {
        fmt.Printf("version: %s\n", version.ClientVersion)
        return
    }

    if *file == "" {
        fmt.Printf("download file is null, quit\n")
        os.Exit(1)
    }

    if *serverAddr != "" && *lookupAddr != "" {
        fmt.Printf("Only can set lookup-tcp-address or server-tcp-address\n")
        os.Exit(1)
    }
    if serverAddr != nil {
        opt.ServerAddr = *serverAddr
    }
    if lookupAddr != nil {
        opt.LookupAddr = *lookupAddr
    }
    if opt.ServerAddr == "" && opt.LookupAddr == "" {
        fmt.Printf("Must set server-tcp-address or lookup-tcp-address\n")
        os.Exit(1)
    }
    opt.MaxWait    = *maxWait
    opt.RoutingNum = *routings
    opt.BlockSize  = *blockSize
    files := strings.TrimRight(*file, " ")
    opt.Files = strings.Split(files, ",")
    opt.SaveDir = *saveDir
    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

    cli := client.NewClient(opt)
    cli.Main()

    <- signalChan
    cli.Exit()
}
