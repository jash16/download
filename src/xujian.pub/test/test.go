package main


import "crypto/md5"
import "fmt"
import (
    "os"
    "bytes"
    "xujian.pub/common"
)

func main() {
    data := []byte("hello xujian")
    md5data := fmt.Sprintf("%x\n", md5.Sum(data))
    fmt.Printf("%s, %d\n", md5data, len(md5data))
    var buf bytes.Buffer
    buf.Reset()
    buf.Write([]byte("hello, world"))
    fmt.Printf("%s\n", buf.String())
    datas := buf.Bytes()
    fmt.Printf("%s\n", datas)
    println(os.Args[1])
    md5s, _ := common.CalFileMd5(os.Args[1])
    fmt.Printf("%s\n", md5s)
    fileSize, _ := common.GetFileSize(os.Args[1])
    fmt.Printf("%d\n", fileSize)
}
