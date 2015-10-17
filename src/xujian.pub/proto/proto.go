package proto

import (
    "net"
    "io"
    "encoding/binary"
)

const (
    FILE int = 0
    SEG  int = 1
    OK   int = 2
    MagicError int = 3
)

type ProtoHeader struct {
    start int64
    end   int64
    md5   string
}

type ProtoType struct {
    typ int
    length int64
    data []byte
}

type ClientReq struct {
    SliceLen int64
    ReqFile  string
    clientId string
}

type Protocol interface {
    IOLoop(conn net.Conn) error
}

func SendResponse(w io.Writer, data []byte) (int, error) {
    //先写数据的长度
    err := binary.Write(w, binary.BigEndian, int32(len(data)))
    if err != nil {
        return 0, err
    }
    //写数据
    n, err := w.Write(data)
    if err != nil {
        return 0, err
    }

    return (n+4), nil
}

func SendFrameResponse(w io.Writer, frameType uint32, data []byte)(int, error) {
    beBuf := make([]byte, 4)
    size := uint32(len(data)) + 4

    binary.BigEndian.PutUint32(beBuf, size)
    n, err := w.Write(beBuf)
    if err != nil {
        return 0, err
    }

    binary.BigEndian.PutUint32(beBuf, frameType)
    n, err = w.Write(beBuf)
    if err != nil {
        return 0, err
    }

    n, err = w.Write(data)
    if err != nil {
        return 0, err
    }
    return n + 8, nil
}

func ReadResponse(r io.Reader) ([]byte, err) {
    var msgSize int32

    err := binary.Read(r, binary.BigEndian, &msgSize)
    if err != nil {
        return nil, err
    }

    buf := make([]byte, msgSize)

    _, err := io.ReadFull(r, buf)
    if err != nil {
        return nil, err
    }
    return buf, err
}
