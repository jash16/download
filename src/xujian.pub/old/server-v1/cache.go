package server

import (
    "sync"
    "fmt"
    "time"
    "errors"
    "strings"
)

const (
    defaultCacheSizePerBlock int64 = 1048576 // 1M
)

type Block struct {
    StartTime int64
    LastAccess int64

    Hits     int
    StartPos int64
    EndPos   int64
    Data     []byte
}


type Cache struct {
    Hits int64
    MisHits int64
    maxCacheBlock int64
    Blocks map[string]*Block //key is filename_start_end
    sync.RWMutex
}

func NewBlock() *Block{
    return &Block {
        Hits: 0,
        StartPos: 0,
        EndPos: 0,
    }
}

func NewCache(maxBlocks int64) *Cache {
    return &Cache {
        Blocks: make(map[string]*Block),
        Hits: 0,
        maxCacheBlock: maxBlocks,
    }
}

func (c *Cache) LookupBlock(file string, start, end int64) *Block{
    s, e := c.adjustPos(start, end)
    key := fmt.Sprintf("%s_%d_%d", file, s, e)
    c.RLock()
    if b, ok := c.Blocks[key]; ok {
        c.RUnlock()
        c.Hits ++
        return b
    }
    c.RUnlock()
    c.MisHits ++
    return nil
}

func (c *Cache)adjustPos(start, end int64) (startPos, endPos int64) {
    s, e := int64(0), int64(0)
    if start % defaultCacheSizePerBlock != 0 {
        s = start
    } else {
        s = start - start % defaultCacheSizePerBlock
    }

    if end % defaultCacheSizePerBlock != 0 {
        e = end
    } else {
        e = end + defaultCacheSizePerBlock - end % defaultCacheSizePerBlock
    }

    return s, e
}

func (c *Cache)AddOrHitCache(file string, data []byte, start, end int64) error {

    b := c.LookupBlock(file, start, end)
    if b != nil {
        b.Hits ++
        b.LastAccess = time.Now().Unix()
        return nil
    }

    b = NewBlock()
    b.SetData(data, start, end)

    s, e := c.adjustPos(start, end)
    if s != start || end != e {
        return errors.New("start or end do not same as the ajustPos start and end")
    }
    key := fmt.Sprintf("%s_%d_%d\n", file, s, e)
    c.AddBlock(key, b)
    return nil
}

func (c *Cache)AddBlock(key string, b *Block) {
    c.Lock()
    c.Blocks[key] = b
    c.Unlock()
}

func (c *Cache)RemoveBlock(key string, b *Block) {
    c.Lock()
    delete(c.Blocks, key)
    c.Unlock()
}

func (c *Cache)RemoveBlockByKey(file string) {
    c.Lock()
    for k, _ := range c.Blocks {
         if strings.HasPrefix(k, file + "_") {
             delete(c.Blocks, k)
         }
    }
    c.Unlock()
}

//you should keep start < end 
func (b *Block)ReadRata(start, end int64) []byte {
    if (start >= b.StartPos && end <= b.EndPos) {
        return b.Data[start:end]
    }
    return nil
}

func (b *Block)SetData(data []byte, start, end int64) {
    b.Data = make([]byte, end - start)
    b.Hits = 1
    b.StartPos = start
    b.EndPos   = end
    b.StartTime , b.LastAccess = time.Now().Unix(), time.Now().Unix()
    copy(b.Data, data)
    return
}
