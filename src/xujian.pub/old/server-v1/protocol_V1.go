package server

import (
    "os"
    "io"
    "net"
    "fmt"
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
            //close or continue
            p.sendData(client, []byte(err.Error()))
            p.ctx.s.logf("get error: %s", []byte(err.Error()))
            break
        }
    }
    client.conn.Close()
    client.ctx.s.logf("client[addr: %s] quit now", client.Addr)
    return err
    //new client
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

    p.ctx.s.logf("get requeset for file: %s", fullFile)
    md5s, err := common.CalFileMd5(fullFile)
    if err != nil {
        return err
    }
    fileSize, err := common.GetFileSize(fullFile)
    if err != nil {
        return err
    }
    sendData := [][]byte{[]byte("SEND"), []byte(md5s), []byte(fmt.Sprintf("%d", fileSize))}
    //client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
    var buf bytes.Buffer
    for _, value := range sendData {
        buf.Write(value)
        buf.Write(separatorBytes)
    }
    buf.Write(newLineBytes)
    _, err = client.Writer.Write(buf.Bytes())
    if err != nil {
        return err
    }
    client.Writer.Flush()
    return nil
}

func (p *protocolV1) processBlockRequest (client *Client, file string, start, end int64) {
    client.Lock()
    //var fileInfo *FileInfo
    if _, ok := client.Files[file]; ok == false {
        fullFile := p.ctx.s.Opts.DataPath + file
        fin, err := os.OpenFile(fullFile, os.O_RDONLY, 0666)
        if err != nil {
            client.Unlock()
            err2 := p.ProcessErr("E_NOT_EXIST", fmt.Sprintf("file: %s not exist", file))
            p.sendData(client, []byte(err2.Error()))
            return
        }
        fileInfo := &FileInfo {
            Fhandler: fin,
            FileName: file,
        }
        client.Files[file] = fileInfo
    }
    fileInfo := client.Files[file]
    client.Unlock()

    fileInfo.Lock()
    defer fileInfo.Unlock()
    f := fileInfo.Fhandler
    _, err := f.Seek(start, 0)
    if err != nil {
        err2 := p.ProcessErr("E_SEEK_ERROR", fmt.Sprintf("seek error: %s", err.Error()))
        p.sendData(client, []byte(err2.Error()))
        return
    }
    buf := make([]byte, end - start + 1)
    _, err = io.ReadFull(f, buf)
    if err != nil {
        err2 := p.ProcessErr("E_READ_ERROR", fmt.Sprintf("read error: %s", err.Error()))
        p.sendData(client, []byte(err2.Error()))
        return
    }
    err = p.sendBlock(client, buf)
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
    p.sendData(client, buf.Bytes())
    //buf.Write(data)
/*
    client.Writer.Write(buf.Bytes())
    client.ctx.s.logf("send block meta: %s", buf.Bytes())

    client.Writer.Write(data)
    client.ctx.s.logf("send block data: %s", data)
    //client.conn.SetWriteDeadline(time.Now().Add(100 * time.Second))

    client.Writer.Flush()
    */
    return p.sendData(client, data)
}

func (p *protocolV1) sendData(client *Client, data []byte) error {
    _, err := client.Writer.Write(data)
    if err != nil {
        return err
    }
    client.Writer.Flush()
    return nil
}
