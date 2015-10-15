package server

import (
    "xujian.pub/proto"
    "xujian.pub/common"
    "xujian.pub/common/util"
    //"xujian.pub/client"
    "time"
    "os"
    "net"
    "sync"
    "fmt"
)

type Server struct {
    Opts *ServerOption
    sync.RWMutex

    err error
    Clients map[string]*Client //clients

    notifyChan chan interface{}
    exitChan   chan int

    //stat info
    ClientNum int64
    CurClientNum int64
    DownloadNum int64
    DownloadSize int64

    md5Lock sync.RWMutex
    //md5Cache map[string]*MetaInfo
    md5Cache MetaCache

    cacheLock sync.RWMutex
    dataCache *Cache

    cacheEventChan chan *cacheEvent

    wg   util.WaitGroupWrapper
    tcpListener net.Listener
}

func NewServer(opts *ServerOption) *Server {
    s := &Server {
        Opts:  opts,
        Clients: make(map[string]*Client),
        notifyChan: make(chan interface{}),
        exitChan: make(chan int),
        cacheEventChan: make(chan *cacheEvent),
        ClientNum: 0,
        DownloadNum: 0,
        DownloadSize: 0,
        md5Cache: make(map[string]*MetaInfo),
        dataCache: NewCache(1000),
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

    //watch file, to change the cache
    s.wg.Wrap(func() {
        s.watchLoop()
    })
}

func (s *Server) cacheLoop() {
    s.logf("cacheLoop start")
    var exitFlag bool = false
    expireCache := time.NewTicker(30 * time.Second)
    for {
        select {
        case <- expireCache.C:
            //
        case cev := <- s.cacheEventChan:
            time.Sleep(1 * time.Second)
            s.md5Lock.Lock()
            s.md5Cache.ProcessEvent(cev)
            s.md5Lock.Unlock()

            s.cacheLock.Lock()
            s.dataCache.ProcessEvent(cev)
            s.cacheLock.Unlock()
        //process this file, md5, file content
        case <- s.exitChan:
            exitFlag = true
        }

        if exitFlag == true {
            break
        }
    }
    s.logf("cacheLoop stop")
}

func (s *Server) statLoop() {
    s.logf("statLoop start")
    var exitFlag bool = false

    t := time.NewTicker(10 * time.Second)
    for {
        select {
        case <- t.C:
            s.RLock()
            s.logf("total_clients: %d, current_clients: %d, download_num: %d, download_size: %d",
                   s.ClientNum, s.CurClientNum, s.DownloadNum, s.DownloadSize)
            s.logf("cache_hits: %d, cache_miss: %d\n", s.dataCache.Hits, s.dataCache.MisHits)
            s.RUnlock()
        case <- s.exitChan:
            exitFlag = true
        }

        if exitFlag == true {
            break
        }
    }

    s.logf("statLoop exit")
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

func (s *Server) GetCache(file string) *FileBlock {
    s.cacheLock.RLock()
    defer s.cacheLock.RUnlock()

    cache := s.dataCache.LookupFile(file)

    return cache
}

//check and delete some cache
func (s *Server) AddDataCache(file string, data []byte) {
    s.dataCache.AddOrHitCache(file, data)
    return
}

func (s *Server) DeleteDataCache(file string) {
    s.cacheLock.Lock()
    defer s.cacheLock.Unlock()
    s.dataCache.RemoveBlockByKey(file)
    //delete(s.dataCache, file)
}

func (s *Server)GetMetaInfo(file string) *MetaInfo {
    s.md5Lock.RLock()
    defer s.md5Lock.RUnlock()

    metaInfo, ok := s.md5Cache[file]
    if ok {
        return metaInfo
    }
    return nil
}

func (s *Server)AddMetaCache(file string, metaInfo *MetaInfo) {
    s.md5Lock.Lock()
    defer s.md5Lock.Unlock()
    _, ok := s.md5Cache[file]
    if ok {
        return
    }
    s.md5Cache[file] = metaInfo
}

func (s *Server) DeleteMetaCache(file string) {
    s.md5Lock.Lock()
    defer s.md5Lock.Unlock()
    delete(s.md5Cache, file)
}
