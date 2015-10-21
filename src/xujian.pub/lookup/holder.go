package lookup

import (
    "time"
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
    paused bool //the download server can stop work on his own
    pasuedTime time.Time
    peerInfo *PeerInfo
}

type Producers []*Producer

func NewHolderDB() *HolderDB {
    return &HolderDB {
        HolderMap: make(map[string]Producers)
    }
}

func (h *HolderDB) AddProducer(file string, peerInfo *PeerInfo) {
    h.Lock()
    defer h.Unlock()

    var found bool
    producer := &Producer{
        peerInfo: peerInfo
        pasued: false,
        pasuedTime:nil
    }
    producers, ok := h.HoldMap[file]
    if ok {
        for _, p := range(producers) {
            if peerInfo.id == p.peerInfo.id {
                found = true
            }
        }
        if found == false {
            h.HoldMap[file] = append(producers, producer)
        }
    } else {
        producers := Producers{}
        h.HoldMap[file] = append(producers, producer)
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

func (h *HolderDB) FindProperProducer(file string) *Producer{
    h.RLock()
    defer h.RUnlock()

    if producers, ok := h.HoldMap[file]; ok {
        if len(producers) > 0 {
            return producers[0]
        }
    }
    return nil
}

func (h *HolderDB) RemoveFileProducer(file string, peerInfo *PeerInfo) error {
    h.Lock()
    defer h.Unlock()

    var found bool
    cleaned := Producers{}
    producers, ok := h.HoldMap[file]
    if ! ok {
        return fmt.Errorf("E_NOT_FOUND")
    }
    for _, producer := range(producers) {
        if peerInfo.id != producer.peerInfo.id {
            cleaned = append(cleaned, producer)
        } else {
            found = true
        }
    }

    if found == true {
        s.HoldMap[file] = cleaned
    }

    return nil
}

func (h *HolderDB) RemoveProducer(peerInfo *PeerInfo) error {
    h.Lock()
    defer h.Unlock()

    cleaned := Producers{}

    for file, producers := range(h.HoldMap) {
        for _, producer := range(producers) {
            if peerInfo.id != producer.peerInfo.id {
                cleaned = append(cleaned, producer)
            }
        }

        h.HoldMap[file] = cleaned
    }
    return nil
}


