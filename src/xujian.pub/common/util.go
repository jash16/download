package common

import (
    "os"
    "fmt"
    "io/ioutil"
    "crypto/md5"
)

func IsFileExist(path string) bool {
    _, err := os.Stat(path)
    return err == nil || os.IsExist(err)
}

func CalMd5(data []byte) string {
    return fmt.Sprintf("%x", md5.Sum(data))
}

func CalFileMd5(file string) (string, error) {
    //fin, err := os.OpenFile(file, os.O_RDONLY, 0644)
    //if err != nil {
    //    return "", err
    //}
    //defer fin.Close()

    buf, err := ioutil.ReadFile(file)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%x", md5.Sum(buf)), nil
}

func GetFileSize(file string) (int64, error) {
    f, err := os.Open(file)
    if err != nil {
        return 0, err
    }
    defer f.Close()
    statInfo, err := f.Stat()
    if err != nil {
        return 0, err
    }
    return statInfo.Size(), nil
}

func GenerateId(file string) string {
    return ""
}
