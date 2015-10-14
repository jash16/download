package client

import (
    "net"
    "io"
    "sync"
    "xujian.pub/common"
    "fmt"
    "sync/atomic"
    "bytes"
    "bufio"
    "time"
    "strconv"
    "os"
)

type reloadPkg struct {
    filename string
    retry int
    startPos int64
    endPos   int64
}

type downConn struct {
    conn *net.TCPConn
    r io.Reader
    w io.Writer
}

type worker struct {
    reqConn *net.TCPConn
    r io.Reader
    w io.Writer
    file string
    sync.Mutex
    fhandler *os.File
    fInfo *fileInfo
    client *Client
    conns []*downConn
    reloadChan chan reloadPkg
}

func NewdownConn(client *Client) *downConn {
    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
    }
    conn, err := dialer.Dial("tcp", client.Opts.ServerAddr)
    if err != nil {
        client.logf("[FATAL] client dial: %s failed, quit now", client.Opts.ServerAddr)
        return nil
    }
    tcpConn := conn.(*net.TCPConn)
    return &downConn {
        conn: tcpConn,
        r: tcpConn,
        w: tcpConn,
    }
}

func Newworker(file string, client *Client) *worker{
    client.logf("new workers")
    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
    }
    conn, err := dialer.Dial("tcp", client.Opts.ServerAddr)
    if err != nil {
       client.logf("[FATAL] client dial: %s failed, quit now", client.Opts.ServerAddr)
       os.Exit(1)
    }
    tcpConn := conn.(*net.TCPConn)

    return &worker {
        reqConn: tcpConn,
        r: tcpConn,
        w: tcpConn,
        file: file,
        client: client,
        reloadChan: make(chan reloadPkg),
        conns: make([]*downConn, client.Opts.RoutingNum),
    }
}

func (w *worker) StartDownload() {
    var needDown int64
    var fileSize int64
    var blockNum int64
    var routingNum int64
    var blockNumPerRouting int64
    var totalRouting int64
    var i int64
    var dconn *downConn
    var t *time.Ticker
    _, err := w.reqConn.Write(common.MagicV1)
    if err != nil {
        w.reqConn.Close()
        w.client.logf("[FATAL] write magic failed, %s", err.Error())
        os.Exit(1)
    }

    //send requset file
    err = w.sendRequest()
    if err != nil {
        w.client.logf("send request failed, error: %s\n", err.Error())
    //    return
        goto END
    }

    fileSize = w.fInfo.fileSize
    blockNum = fileSize / w.client.Opts.BlockSize
    routingNum = w.client.Opts.RoutingNum
    blockNumPerRouting = int64(blockNum / routingNum)
    if blockNumPerRouting > 1 {
        totalRouting = routingNum
    } else {
        totalRouting = blockNum
    }
    needDown = blockNum + 1
    w.client.logf("total thread: %d", totalRouting)
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

        w.client.wg.Wrap(func(){
            dconn := NewdownConn(w.client)
            if dconn == nil {
                w.client.logf("new downConn failed")
                os.Exit(1)
            }
            _, err := dconn.w.Write(common.MagicV1)
            if err != nil {
                dconn.conn.Close()
                w.client.logf("[FATAL] write magic failed, %s", err.Error())
                os.Exit(1)
            }
            w.download(dconn, start, end, w.client.Opts.BlockSize)
            atomic.AddInt64(&needDown, -1)
        })
    }

    dconn = &downConn {
        conn: w.reqConn,
        r: w.r,
        w: w.w,
    }
    w.client.wg.Wrap(func() {
        startPos := blockNum * w.client.Opts.BlockSize
        endPos   := fileSize
        w.downloadFileRange(dconn, startPos, endPos)
        atomic.AddInt64(&needDown, -1)
    })

END:
    t = time.NewTicker(1 * time.Second)
    for {
         select {
         case pkg := <- w.reloadChan:
             w.client.wg.Wrap(func() {
                 atomic.AddInt64(&needDown, 1)
                 w.reload(dconn, pkg)
                 atomic.AddInt64(&needDown, -1)
             })
         case <- t.C:
         }
         if atomic.LoadInt64(&needDown) == 0 {
             w.client.doneChan <-true
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
func (w *worker) download(dconn *downConn, start, end int64, blockSize int64) {
    for i := start; i < end; i ++ {
        startPos := i * blockSize
        endPos   := startPos + blockSize
        err := w.downloadFileRange(dconn, startPos, endPos)
        if err != nil {
            pkg := reloadPkg{
                startPos: startPos,
                endPos: endPos,
            }
            w.reloadChan <- pkg
        }
    }
}

func (w *worker)reload(conn *downConn, pkg reloadPkg) {
    err := w.downloadFileRange(conn, pkg.startPos, pkg.endPos)
    if err != nil {
        w.client.logf("reload failed")
    }
}

/*
    @param:
        startPos: download file pos
        endPos  : end position of the file
*/
func (w *worker) downloadFileRange(dconn *downConn, startPos, endPos int64) error {
    // download from startPos to endPos
    reqString := [][]byte{[]byte("GET"), []byte(w.file), []byte("DATA"), []byte(fmt.Sprintf("%d", startPos)), []byte(fmt.Sprintf("%d", endPos - 1))}

    var buf bytes.Buffer

    for _, value := range reqString {
        buf.Write([]byte(value))
        buf.Write([]byte(" "))
    }
    buf.Write([]byte("\n"))
    _, err := dconn.w.Write(buf.Bytes())
    if err != nil {
        return err
    }

    reader := bufio.NewReaderSize(dconn.r, 1048576+1024)
    line, err := reader.ReadSlice('\n')
    if err != nil {
        return err
    }
    println("line length: ",len(line))
    line = bytes.TrimRight(line, "\n")
    line = bytes.TrimRight(line, "\r")
    line = bytes.TrimRight(line, " ")
    bInfoBytes := bytes.Split(line, []byte(" "))
    if len(bInfoBytes) < 3 {
        w.client.logf("download data: need more meta info of block")
        return fmt.Errorf("download data: need more meta info of block")
    }
    if "SEND" != string(bInfoBytes[0]) {
        w.client.logf("download data: server respone error")
        return fmt.Errorf("download data: server respone error")
    }
    md5 := string(bInfoBytes[1])
    length,_ := strconv.Atoi(string(bInfoBytes[2]))

    //println("data length",length)
    data := make([]byte, length)
    //_, err = reader.Read(data)
    _, err = io.ReadFull(reader, data)
    if err != nil {
        return err
    }

    //w.client.logf("receive: %s\n", data)
    // cal md5
    md5Recv := common.CalMd5(data)
    if md5Recv != md5 {
        w.client.logf("cal md5(%s) not same as the recevie md5(%s)", md5Recv, md5)
        return fmt.Errorf("cal md5 not same as the recevie md5")
    }

    // write to file
    return w.writeToFile(startPos, data)
}

func (w *worker) sendRequest() error {
    var buf bytes.Buffer
    reqString := []string{"GET", w.file, "META"}
    for _, value := range reqString {
        buf.Write([]byte(value))
        buf.Write([]byte(" "))
    }
    buf.Write([]byte("\n"))
    w.client.logf("send request data: %s", buf.Bytes())
    err := w.Write(buf.Bytes())
    if err != nil {
        return err
    }

    w.client.logf("wait for meta data")
    buf2 := make([]byte, 128)
    _, err = w.Read(buf2)
    if err != nil {
        return err
    }

    err = w.parseMetaInfo(buf2)
    if err != nil {
        return err
    }
    return nil
}

func (w *worker) Write(data []byte) error {
    w.reqConn.SetWriteDeadline(time.Now().Add(5*time.Second))
    _, err := w.w.Write(data)
    return err
}

func (w *worker) Read(data []byte) (int, error) {
    w.reqConn.SetReadDeadline(time.Now().Add(10*time.Second))
    return w.r.Read(data)
}

func (w *worker) parseMetaInfo(data []byte) error {
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
    w.client.logf("file: %s, md5: %s, size: %d", w.file, md5String, fileSize)
    fInfo := NewfileInfo(md5String, int64(fileSize), w.file)
    w.fInfo = fInfo
    return nil
}

func (w *worker) writeToFile(seekPos int64, data []byte) error {
    w.Lock()
    defer w.Unlock()
    if w.fhandler == nil {
        fullFile := w.client.Opts.SaveDir + w.file
        f, err := os.Create(fullFile)
        if err != nil {
            return err
        }
        w.fhandler = f
        err = w.fhandler.Truncate(w.fInfo.fileSize)
        if err != nil {
            return err
        }
    }
    _, err := w.fhandler.Seek(seekPos, 0)
    if err != nil {
        return err
    }
    _, err = w.fhandler.Write(data)
    if err != nil {
        return err
    }
    return nil
}
