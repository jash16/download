package common

import (
    "io"
    "strings"
    "encoding/json"
    "encoding/binary"
)

type Command struct {
    Name []byte
    Param [][]byte
    Body []byte
}

var (
    byteSpace = []byte(" ")
    byteNewLine = []byte("\n")
)

func (c *Command) WriteTo(w io.Writer) (int64, error) {
    var total int64
    var buf [4]byte

    n, err := w.Write(c.Name)
    total += int64(n)
    if err != nil {
        return total, err
    }

    for _, param := range(c.Param) {
        n, err := w.Write(byteSpace)
        total += int64(n)
        if err != nil {
            return total, err
        }
        n, err = w.Write(param)
        total += int64(n)
        if err != nil {
            return total, err
        }
    }

    n, err = w.Write(byteNewLine)
    total += int64(n)
    if err != nil {
        return total, err
    }

    if c.Body != nil {
        bodyLen := len(c.Body)
        bufs := buf[:]
        binary.BigEndian.PutUint32(bufs, uint32(bodyLen))
        n, err := w.Write(bufs)
        total += int64(n)
        if err != nil {
            return total, err
        }
        n, err = w.Write(c.Body)
        if err != nil {
            return total, err
        }
    }
    return total, nil

}

func Ping() *Command {
    return &Command {[]byte("PING"), nil, nil}
}

func Identify(js map[string]interface{}) (*Command, error) {
    body, err := json.Marshal(js)
    if err != nil {
        return nil, err
    }
    return &Command{[]byte("IDENTIFY"), nil, body}, nil
}

func Register(files []string)(*Command) {
    var fileBuf []byte
    var fileConcat string
    fileConcat = strings.Join(files, "+")
    fileBuf = []byte(fileConcat)

    return &Command{[]byte("REGISTER"), nil, fileBuf}
}

func UnRegister(files []string)(*Command) {

    var fileBuf []byte
    var fileConcat string

    fileConcat = strings.Join(files, "+")
    fileBuf = []byte(fileConcat)
    return &Command{[]byte("UNREGISTER"), nil, fileBuf}
}

func Load(load []byte) *Command {

    return &Command{[]byte("LOAD"), nil, load}
}

func Lookup(file string) *Command {
    return &Command{[]byte("LOOKUP"), [][]byte{[]byte(file)}, nil}
}
