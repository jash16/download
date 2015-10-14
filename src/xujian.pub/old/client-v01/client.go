package client

import (
    "io"
    "net"
    "os"
    "fmt"
    "sync"
    "sync/atomic"
    "bytes"
    "bufio"
    "time"
    "strconv"
    "xujian.pub/common"
    "xujian.pub/common/util"
)

type Client struct {
    Opts *ClientOption
    //serverAddr string
    //downFile    string
    sync.Mutex
    conn *net.TCPConn
    r io.Reader
    w io.Writer
    totalDownload int32
    doneChan chan bool
    fhandler *os.File
    filesInfo map[string]*fileInfo
    exitChan chan bool
    wg util.WaitGroupWrapper
}

type reloadPkg struct {
    filename string
    retry int
    startPos int64
    endPos   int64
}

const (
    defaultReadSize int64 = 1024*1024 + 128 //block size + meta info
)

func NewClient(opt *ClientOption) *Client {
    return &Client {
        Opts: opt,
        filesInfo: make(map[string]*fileInfo),
        exitChan: make(chan bool),
        doneChan: make(chan bool),
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

    dialer := &net.Dialer{
        Timeout:   5 * time.Second,
    }
    conn, err := dialer.Dial("tcp", c.Opts.ServerAddr)
    if err != nil {
        c.logf("[FATAL] client dial: %s failed, quit now", c.Opts.ServerAddr)
        os.Exit(1)
    }
    c.conn = conn.(*net.TCPConn)
    c.r    = conn
    c.w    = conn

    _, err = c.w.Write(common.MagicV1)
    if err != nil {
        c.conn.Close()
        c.logf("[FATAL] write magic failed, %s", err.Error())
        os.Exit(1)
    }

    for _, file := range c.Opts.Files {
        c.wg.Wrap(func() {
            atomic.AddInt32(&c.totalDownload, 1)
            c.StartDownload(file)
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

func (c *Client) StartDownload(file string) {

    var needDown int64
    var reloadChan chan reloadPkg
    reloadChan = make(chan reloadPkg)
    //send requset file
    err := c.sendRequest(file)
    if err != nil {
        c.logf("send request failed, error: %s\n", err.Error())
        return
    }

    if _, ok := c.filesInfo[file]; ok == false {
        c.logf("get failed info error\n")
        return
    }

    fileSize := c.filesInfo[file].fileSize
    blockNum := fileSize / c.Opts.BlockSize
    routingNum := c.Opts.RoutingNum
    blockNumPerRouting := int64(blockNum / routingNum)
    var totalRouting int64
    if blockNumPerRouting > 1 {
        totalRouting = routingNum
    } else {
        totalRouting = blockNum
    }
    var i int64
    needDown = blockNum + 1
    c.logf("total thread: %d", totalRouting)
    for i = 0; i < totalRouting; i ++ {
        var start int64 = 0
        var end int64 = 0
        if i != totalRouting - 1 {
            start = i * blockNumPerRouting
            end   = start + blockNumPerRouting
        } else {
            start = i * blockNumPerRouting
            end = blockNum - start
        }

        c.wg.Wrap(func(){
            c.download(file, start, end, c.Opts.BlockSize)
            atomic.AddInt64(&needDown, -1)
        })
    }

    c.wg.Wrap(func() {
        startPos := blockNum * c.Opts.BlockSize
        endPos   := fileSize
        c.downloadFileRange(file, startPos, endPos)
        atomic.AddInt64(&needDown, -1)
    })

    t := time.NewTicker(1 * time.Second)
    for {
         select {
         case pkg := <- reloadChan:
             c.wg.Wrap(func() {
                 atomic.AddInt64(&needDown, 1)
                 c.reload(pkg)
                 atomic.AddInt64(&needDown, -1)
             })
         case <- t.C:
         }
         if atomic.LoadInt64(&needDown) == 0 {
             c.doneChan <-true
             break
         }
    }
}

/*
    @param:
         file:  file need to download
         start: first block to get
         end:   last block to get
*/
func (c *Client) download(file string, start, end int64, blockSize int64) {

    for i := start; i < end; i ++ {
        startPos := i * blockSize
        endPos   := startPos + blockSize
        c.downloadFileRange(file, startPos, endPos)
    }
}

func (c *Client)reload(pkg reloadPkg) {

}
/*
    @param:
        startPos: download file pos
        endPos  : end position of the file
*/
func (c *Client) downloadFileRange(file string, startPos, endPos int64) error {
    // download from startPos to endPos
    reqString := [][]byte{[]byte("GET"), []byte(file), []byte("DATA"), []byte(fmt.Sprintf("%d", startPos)), []byte(fmt.Sprintf("%d", endPos - 1))}

    var buf bytes.Buffer

    for _, value := range reqString {
        buf.Write([]byte(value))
        buf.Write([]byte(" "))
    }
    buf.Write([]byte("\n"))
    _, err := c.w.Write(buf.Bytes())
    if err != nil {
        return err
    }

    reader := bufio.NewReaderSize(c.r, 128)
    line, err := reader.ReadSlice('\n')
    if err != nil {
        return err
    }
    line = bytes.TrimRight(line, "\n")
    line = bytes.TrimRight(line, "\r")
    line = bytes.TrimRight(line, " ")
    bInfoBytes := bytes.Split(line, []byte(" "))
    if len(bInfoBytes) < 3 {
        c.logf("download data: need more meta info of block")
        return fmt.Errorf("download data: need more meta info of block")
    }
    if "SEND" != string(bInfoBytes[0]) {
        c.logf("download data: server respone error")
        return fmt.Errorf("download data: server respone error")
    }
    md5 := string(bInfoBytes[1])
    length,_ := strconv.Atoi(string(bInfoBytes[2]))

    data := make([]byte, length)
    //_, err = io.ReadFull(c.r, data)
//    _, err = c.r.Read(data)
    _, err = reader.Read(data)
    if err != nil {
        return err
    }

    // cal md5
    md5Recv := common.CalMd5(data)
    if md5Recv != md5 {
        c.logf("cal md5 not same as the recevie md5")
        return fmt.Errorf("cal md5 not same as the recevie md5")
    }

    //c.logf("file: %s, receive: %s", c.Opts.File, string(data))
    // write to file
    return c.writeToFile(file, startPos, data)
}

func (c *Client) sendRequest(file string) error {
    var buf bytes.Buffer
    reqString := []string{"GET", file, "META"}
    for _, value := range reqString {
        buf.Write([]byte(value))
        buf.Write([]byte(" "))
    }
    buf.Write([]byte("\n"))
    c.logf("send request data: %s", buf.Bytes())
    err := c.Write(buf.Bytes())
    if err != nil {
        return err
    }

    c.logf("wait for meta data")
    buf2 := make([]byte, 128)
    _, err = c.Read(buf2)
    if err != nil {
        return err
    }

    err = c.parseMetaInfo(file, buf2)
    if err != nil {
        return err
    }
    return nil
}

func (c *Client) Write(data []byte) error {
    c.conn.SetWriteDeadline(time.Now().Add(5*time.Second))
    _, err := c.w.Write(data)
    return err
}

func (c *Client) Read(data []byte) (int, error) {
    c.conn.SetReadDeadline(time.Now().Add(10*time.Second))
    return c.r.Read(data)
}

func (c *Client) parseMetaInfo(file string, data []byte) error {
    buf := bytes.NewBuffer(data)
    reader := bufio.NewReader(buf)
    info, err := reader.ReadSlice('\n')
    if err != nil {
        return err
    }

//    info = bytes.TrimRight(info, "\n")
//    info = bytes.TrimRight(info, "\r")
    info = info[:len(info)-1]
    if len(info) > 0 && info[len(info)-1] == '\r' {
        info = info[:len(info)-1]
    }
    info = bytes.TrimRight(info, " ")
    metaInfo := bytes.Split(info, []byte(" "))
    if len(metaInfo) != 3 {
        return fmt.Errorf("meta info short some of message")
    }

    fileSize, err := strconv.Atoi(string(metaInfo[2]))
    md5String := string(metaInfo[1])
    c.logf("file: %s, md5: %s, size: %d", file, md5String, fileSize)
    fInfo := NewfileInfo(md5String, int64(fileSize), file)
    c.filesInfo[file] = fInfo
    return nil
}

func (c *Client) writeToFile(file string, seekPos int64, data []byte) error {
    c.Lock()
    defer c.Unlock()
    if c.fhandler == nil {
        fullFile := c.Opts.SaveDir + file
        f, err := os.Create(fullFile)
        if err != nil {
            return err
        }
        c.fhandler = f
        err = c.fhandler.Truncate(c.filesInfo[file].fileSize)
        if err != nil {
            return err
        }
    }
    _, err := c.fhandler.Seek(seekPos, 0)
    if err != nil {
        return err
    }
    _, err = c.fhandler.Write(data)
    if err != nil {
        return err
    }
    return nil
}

func (c *Client) Exit() {
    close(c.exitChan)
    c.Lock()
    if c.fhandler != nil {
        c.fhandler.Close()
    }
    c.Unlock()

    c.wg.Wait()
    os.Exit(0)
    return
}
