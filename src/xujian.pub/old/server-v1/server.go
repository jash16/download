package server

import (
    "xujian.pub/proto"
    "xujian.pub/common"
    "xujian.pub/common/util"
    "xujian.pub/client"
    "os"
    "net"
    "sync"
    "fmt"
)

type Server struct {
    Opts *ServerOption
    sync.RWMutex

    err error
    Clients map[string]*client.Client //cliens

    notifyChan chan interface{}
    exitChan   chan int

    //stat info
    ClientNum int64
    DownloadNum int64
    DownloadSize int64

    wg   util.WaitGroupWrapper
    tcpListener net.Listener
}

func NewServer(opts *ServerOption) *Server {
    s := &Server {
        Opts:  opts,
        Clients: make(map[string]*client.Client),
        notifyChan: make(chan interface{}),
        exitChan: make(chan int),
        ClientNum: 0,
        DownloadNum: 0,
        DownloadSize: 0,
    }

    if common.IsFileExist(opts.DataPath) == false {
        s.logf("Fatal: data path %s not exist", opts.DataPath)
        os.Exit(-1)
    }

    return s
}

func (s *Server) logf(f string, args ...interface{}) {
    if s.Opts.Logger == nil {
        return
    }
    s.Opts.Logger.Output(2, fmt.Sprintf(f, args...))
}

func (s *Server) Main() {

    ctx := &context{s}
    //listener
    tcpListener, err := net.Listen("tcp", s.Opts.TCPAddress)
    if err != nil {
        s.logf("FATAL: listen: %s failed - %s", s.Opts.TCPAddress, err)
        os.Exit(1)
    }

    s.Lock()
    s.tcpListener = tcpListener
    s.Unlock()

    //start tcp server
    tcpServer := &tcpServer {ctx: ctx}
    s.wg.Wrap(func() {
        proto.TCPServer(tcpListener, tcpServer, s.Opts.Logger)
    })

    //cache
    s.wg.Wrap(func() {
        s.cacheLoop()
    })
    //stat
    s.wg.Wrap(func() {
        s.statLoop()
    })
}

func (s *Server) cacheLoop() {

}

func (s *Server) statLoop() {

}

func (s *Server) Exit() {
    s.Lock()
    if s.tcpListener != nil {
        s.tcpListener.Close()
    }
    s.Unlock()

    close(s.exitChan)

    s.wg.Wait()
}

