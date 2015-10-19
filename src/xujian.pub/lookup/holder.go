package lookup

import (
    "sync"
)

type HolderDB struct {
    sync.RWMutex
    HoldMap map[string]Producers
}

type PeerInfo struct {
    lastActive int64
    id string

    RemoteAddr string `flag:"remote-address"`
    Hostname string `flag:"hostname"`
    TcpAddress string `flag:"tcp-address"`
    HttpAddress string `flag:"http-address"`
    Version string `flag:"version"`
}

type Producer struct {
    peerInfo *PeerInfo
}

type Producers []*Producer

func NewHolderDB() *HolderDB {
    return &HolderDB {
        HolderMap: make(map[string]Producers)
    }
}

func (h *HolderDB) FindProducers(file string) Producers {
    h.RLock()
    defer h.RUnlock()
    if producers, ok := h.HoldMap[file]; ok {
        return producers
    }
    return nil
}

func (h *HolderDB) FindProperProducer(file string) Producer {
    h.RLock()
    defer h.RUnlock()

    if producers, ok := h.HoldMap[file]; ok {
    
    }
    return nil
}
