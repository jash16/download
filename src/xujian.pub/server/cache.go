package server

import (
    "sync"
    "fmt"
    "xujian.pub/common"
    "time"
    "io/ioutil"
    //"errors"
//    "strings"
)

const (
    defaultCacheSizePerBlock int64 = 1048576 // 1M
    ADD int = 0
    DEL int = 1
    MOD int = 2
    RENAME int = 3
)

type MetaInfo struct {
    FileName string
    FileSize int64
    Md5Info  string
    sync.RWMutex
}

type MetaCache map[string]*MetaInfo

func (m MetaCache) ProcessEvent(cev *cacheEvent) error {
    typ := cev.eventType
    file := cev.filename
    fmt.Printf("get metaInfo modify: %d, %s\n", typ, file)
    switch(typ) {
    case ADD, MOD:
        md5, err := common.CalFileMd5(file)
        if err != nil {
            return err
        }
        fileSize, err := common.GetFileSize(file)
        if err != nil {
            return err
        }
        metaInfo := &MetaInfo {
            FileName: file,
            FileSize: fileSize,
            Md5Info: md5,
        }
        m[file] = metaInfo
        delete(m, file)
    case DEL:
        if _, ok := m[file]; ok {
            delete(m, file)
        }
    }
    return nil
}

type FileBlock struct {
    StartTime int64
    LastAccess int64

    Hits int

    FileSize   int64
    Data     []byte
}


type Cache struct {
    Hits int64
    MisHits int64
    CacheFiles int64
    CacheSize int64

    Expire time.Duration
    maxCacheBlock int64
    Blocks map[string]*FileBlock //key is filename
    sync.RWMutex
}

type cacheEvent struct {
    eventType int
    filename string
}

func NewBlock() *FileBlock{
    return &FileBlock {
        LastAccess: time.Now().Unix(),
        Hits: 0,
        FileSize: 0,
    }
}

func NewCache(maxBlocks int64) *Cache {
    return &Cache {
        Blocks: make(map[string]*FileBlock),
        Hits: 0,
        maxCacheBlock: maxBlocks,
    }
}

func (c *Cache) LookupFile(file string) *FileBlock{
//    c.RLock()
//    defer c.RUnlock()
    if b, ok := c.Blocks[file]; ok {
        c.Hits ++
        return b
    }
    c.MisHits ++
    return nil
}


func (c *Cache)AddOrHitCache(file string, data []byte) error {

    c.Lock()
    defer c.Unlock()
    b := c.LookupFile(file)
    if b != nil {
        b.LastAccess = time.Now().Unix()
        return nil
    }

    b = NewBlock()
    b.SetData(data)

    c.AddBlock(file, b)
    return nil
}

func (c *Cache)AddBlock(key string, b *FileBlock) {
    //c.Lock()
    c.Blocks[key] = b
    c.CacheFiles += 1
    c.CacheSize += b.FileSize
    //c.Unlock()
}

func (c *Cache)ReadData(file string, start, end int64) []byte {
    c.RLock()
    defer c.Unlock()
    b := c.LookupFile(file)
    if b != nil && b.FileSize >= end {
        fmt.Printf("file: %s, hit the cache \n", file)
        data := make([]byte, end-start+1)
        copy(data, b.Data[start:end+1])
        c.Hits ++
        return data
    }
    if b == nil {
        fmt.Printf("b is nil, file: %s\n", file)
    } else {
        fmt.Printf("file size: %d, end=%d\n", b.FileSize, end)
    }
    c.MisHits ++
    return nil
}

func (c *Cache)RemoveBlock(key string, b *FileBlock) {
    c.Lock()
    delete(c.Blocks, key)
    c.Unlock()
}

func (c *Cache) RemoveBlockByKey(file string) {
    c.Lock()
    for k, _ := range c.Blocks {
         //if strings.HasSuffix(k, file) {
         if k == file {
             fmt.Printf("delete file: %s, %s\n", k, file)
             delete(c.Blocks, file)
         }
    }
    c.Unlock()
}

func (c *Cache)ExpireBlock() int64 {
    return int64(0)
}

//指定时间内扫描cache
func (c *Cache)ExpireBlockStep() int64 {
    return int64(0)
}

func (c *Cache) ProcessEvent(ev *cacheEvent) error {
    typ := ev.eventType
    file := ev.filename
    fmt.Printf("get data cache modify: %d, %s\n", typ, ev.filename)
    switch(typ) {
    case MOD:
        data, err := ioutil.ReadFile(file)
        fmt.Printf("=============read event: data %s\n", data)
        if err != nil {
            return err
        }
        c.RemoveBlockByKey(file)
        //c.AddOrHitCache(file, data)
    case DEL:
        c.RemoveBlockByKey(file)
    }
    return nil
}

//you should keep start < end 
func (b *FileBlock)ReadData(start, end int64) []byte {
    fmt.Printf("start: %d, end: %d, size: %d\n", start, end, b.FileSize)
    if (start >= 0 && end <= b.FileSize) {
        data := make([]byte, end-start+1)
        copy(data, b.Data[start:end+1])
        return data
    }
    return nil
}

func (b *FileBlock)SetData(data []byte) {
    fmt.Printf("set data: %s\n", data)
    fileSize := len(data)
    b.Data = make([]byte, fileSize)
    b.FileSize = int64(fileSize)
    b.Hits = 1
    b.StartTime , b.LastAccess = time.Now().Unix(), time.Now().Unix()
    copy(b.Data, data)
    return
}
