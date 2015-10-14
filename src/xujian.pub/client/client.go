package client

import (
    "os"
    "fmt"
    "sync"
    "sync/atomic"
    "xujian.pub/common/util"
)

type Client struct {
    Opts *ClientOption
    sync.Mutex
    workers []*worker
    totalDownload int32
    doneChan chan bool
    exitChan chan bool
    wg util.WaitGroupWrapper
}

const (
    defaultReadSize int64 = 1024*1024 + 1024 //block size + meta info
)

func NewClient(opt *ClientOption) *Client {
    return &Client {
        Opts: opt,
        exitChan: make(chan bool),
        doneChan: make(chan bool),
        workers: make([]*worker, len(opt.Files)),
        totalDownload: 0,
    }
}


func (c *Client) logf(s string, args ...interface{}) {
    if c.Opts.Logger == nil {
        return
    }

    c.Opts.Logger.Output(2, fmt.Sprintf(s, args...))
}

func (c *Client) Main() {
    c.logf("client main")

    for _, file := range c.Opts.Files {
        worker := Newworker(file, c)
        c.workers = append(c.workers, worker)
        c.wg.Wrap(func() {
            atomic.AddInt32(&c.totalDownload, 1)
            worker.StartDownload()
        })
    }

    for {
        select {
        case <- c.doneChan:
            atomic.AddInt32(&c.totalDownload, -1)
        }
        if atomic.LoadInt32(&c.totalDownload) == 0 {
            break
        }
    }
    c.logf("all files have done, quit now")
    c.Exit()
}

func (c *Client) Exit() {
    close(c.exitChan)
    c.wg.Wait()
    os.Exit(0)
    return
}
