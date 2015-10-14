package server

import (
    "os"
    "sync"
)

type FileInfo struct {
    FileName string
    Fhandler *os.File
    sync.Mutex
}
