package lookup

import (
    "os"
    "net"
    "fmt"
    "sync"
    "xujian.pub/proto"
    _ "xujian.pub/common"
    "xujian.pub/common/util"
)

type LookupServer struct {
    Opts *LookupOptions

    sync.Mutex
    tcpListener net.Listener
    httpListener net.Listener

    Hold *HolderDB
    wg util.WaitGroupWrapper

    exitChan chan bool
}

func NewLookupServer(opt *LookupOptions) *LookupServer {
    return &LookupServer {
        Opts: opt,
        exitChan: make(chan bool),
        Hold: NewHolderDB(),
    }
}

func (l *LookupServer) Main() {
    l.logf("lookup server start")

    tcpListener, err := net.Listen("tcp", l.Opts.TcpAddress)
    if err != nil {
        l.logf("error, Listen(%s) failed - %s", l.Opts.TcpAddress, err)
        os.Exit(1)
    }
    l.Lock()
    l.tcpListener = tcpListener
    l.Unlock()

    ctx := &context{s: l}
    tcpSrv := &tcpServer{ctx: ctx}
    l.wg.Wrap(func() {
        proto.TCPServer(l.tcpListener, tcpSrv, l.Opts.Logger)
    })
/*
    httpListener, err := net.Listen("tcp", l.Opts.HttpAddress)
    if err != nil {
        l.logf("error, Listen(%s) failed - %s", l.opts.HttpAddress, err)
        os.Exit(1)
    }
    l.Lock()
    l.httpListener = httpListener
    l.Unlock()
*/
}

func (l *LookupServer) logf(f string, args ...interface{}) {
    if l.Opts.Logger != nil {
        l.Opts.Logger.Output(2, fmt.Sprintf(f, args...))
    }
    return
}

func (l *LookupServer) Exit() {
    l.logf("lookup server quit")

    l.Lock()
    if l.tcpListener != nil {
        l.logf("close tcp listener")
        l.tcpListener.Close()
    }
    l.Unlock()

    l.Lock()
    if l.httpListener != nil {
        l.logf("close http listener")
        l.httpListener.Close()
    }
    l.Unlock()

    close(l.exitChan)

    l.wg.Wait()
}
