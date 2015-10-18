package server

import (
    //"os"
    //"io"
    "net"
    "fmt"
    "io/ioutil"
 //   "time"
    "bytes"
    "strconv"
    "xujian.pub/common"
    "xujian.pub/proto"
)

type protocolV1 struct {
    ctx *context
}

var separatorBytes = []byte(" ")
var newLineBytes = []byte("\n")

const (
    frameError int64 = 0
    frameType  int64 = 0
    frameData  int64 = 1
)

func (p *protocolV1) IOLoop(conn net.Conn) error {
    //process client request
    var err error
    client := NewClient(conn, p.ctx)
    p.ctx.s.Lock()
    p.ctx.s.CurClientNum ++;
    p.ctx.s.Unlock()
    for {
        //client.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
        line, err := client.Reader.ReadSlice('\n')
        if err != nil {
            break
        }
        p.ctx.s.logf("receive data: %s", line)
        line = line[:len(line) - 1]
        if len(line) > 0 && line[len(line) - 1] == '\r' {
            line = line[:len(line) - 1]
        }
        line = bytes.TrimRight(line, " ")
        params := bytes.Split(line, separatorBytes)
        err = p.Exec(client, params)
        if err != nil {
            p.sendData(client, []byte(err.Error()))
            p.ctx.s.logf("get error: %s", []byte(err.Error()))
            break
        }
    }
    client.conn.Close()
    p.ctx.s.Lock()
    p.ctx.s.CurClientNum --
    p.ctx.s.Unlock()
    client.ctx.s.logf("client[addr: %s] quit now", client.Addr)
    return err
}

func (p *protocolV1) Exec(client *Client, params [][]byte) error {
    if "GET" != string(params[0]) {
        return proto.NewFatalClientErr(nil, "E_BAD_PROTOCOL", "bad protocol: " + string(params[0]))
    }
    file := string(params[1])
    //typ := string(params[2])
    switch {
    case bytes.Equal(params[2], []byte("META")):
        return p.SendMetaInfo(client, file)
    case bytes.Equal(params[2], []byte("DATA")):
        start, _ := strconv.Atoi(string(params[3]))
        end, _ := strconv.Atoi(string(params[4]))
        startPos, endPos := int64(start), int64(end)
        p.ctx.s.wg.Wrap(func() {
            p.processBlockRequest(client, file, startPos, endPos)
        })
        return nil
    }
    return nil
}

func (p *protocolV1) SendMetaInfo(client *Client, file string) error {
    fullFile := p.ctx.s.Opts.DataPath + file
    var metaInfo *MetaInfo
    var md5s string
    var fileSize int64
    var err error
    metaInfo = p.ctx.s.GetMetaInfo(file)
    if metaInfo == nil {
        p.ctx.s.logf("get requeset for file: %s", fullFile)
        md5s, err = common.CalFileMd5(fullFile)
        if err != nil {
            return err
        }
        fileSize, err = common.GetFileSize(fullFile)
        if err != nil {
            return err
        }
        metaInfo = &MetaInfo{
            FileName: file,
            FileSize: fileSize,
            Md5Info: md5s,
        }
        p.ctx.s.AddMetaCache(file, metaInfo)
        //p.ctx.s.md5Cache[file] = metaInfo
    } else {
        md5s = metaInfo.Md5Info
        fileSize = metaInfo.FileSize
    }

    sendData := [][]byte{[]byte("SEND"), []byte(md5s), []byte(fmt.Sprintf("%d", fileSize))}

    var buf bytes.Buffer
    for _, value := range sendData {
        buf.Write(value)
        buf.Write(separatorBytes)
    }
    buf.Write(newLineBytes)

    return p.sendData(client, buf.Bytes())
}

func (p *protocolV1) processBlockRequest (client *Client, file string, start, end int64) {
    var cacheHit bool = false
    var data []byte
    var err error
    cache := p.ctx.s.GetCache(file)
    if cache != nil {
        //read data from cache
        p.ctx.s.logf("hit cache, file: %s", file)
        data = cache.ReadData(start, end)
        if data != nil {
            cacheHit = true
        } else {
            p.ctx.s.logf("hit cche, but file: %s data is nil", file)
        }
    }
    if cacheHit == false {
        fullFile := p.ctx.s.Opts.DataPath + file
        data, err = ioutil.ReadFile(fullFile)
        if err != nil {
            p.sendData(client, []byte(err.Error()))
            return
        }
        if len(data) <= 10485760 { // 10M
            p.ctx.s.logf("add file: %s, data: %s, to cache", file, data)
            p.ctx.s.AddDataCache(file, data)
        }
        //p.ctx.s.logf("read data from: %s, data: %s", file, data)
        data = data[start:end+1]
    }

    //p.ctx.s.logf("send data: %d, %s\n", len(data), data)
    err = p.sendBlock(client, data)
    if err != nil {
        p.ctx.s.logf("send block failed: %s", err.Error())
    }
}

func (p *protocolV1) ProcessErr(typ, desc string) error {
    p.ctx.s.logf("[ERROR] %s, %s", typ, desc)
    return fmt.Errorf("%s: %s", typ, desc)
}

func (p *protocolV1) sendBlock(client *Client, data []byte) error {

    var buf bytes.Buffer
    md5 := common.CalMd5(data)
    lenth := fmt.Sprintf("%d", len(data))
    meta := [][]byte{[]byte("SEND"), []byte(md5), []byte(lenth)}
    for _, value := range meta {
        buf.Write(value)
        buf.Write(separatorBytes)
    }
    buf.Write(newLineBytes)

    buf.Write(data)

    return p.sendData(client, buf.Bytes())
}

func (p *protocolV1) sendData(client *Client, data []byte) error {
    _, err := client.Writer.Write(data)
    if err != nil {
        return err
    }
    client.Writer.Flush()
    return nil
}
