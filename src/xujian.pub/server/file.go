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

/*
type MetaInfo struct {
    FileName string
    FileSize int64
    Md5Info  string
    sync.RWMutex
}

type MetaCache map[string]MetaInfo

func (m MetaCache)ProcessMetaInfoChange(typ int, file string) {

}
*/
