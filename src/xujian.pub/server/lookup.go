package server

import (
    "os"
    "time"
    "bytes"
    "encoding/json"
    "xujian.pub/common"
    "xujian.pub/common/version"
)

type FileChange struct {
    typ int32
    file string
}

func (s *Server)lookupLoop() {

    syncLocalFileChan := make(chan *lookupPeer)
    hostname, err := os.Hostname()
    if err != nil {
        s.logf("Fatal: failed to get hostname")
        os.Exit(1)
    }
    for _, addr := range(s.Opts.LookupSrvAddrs) {
        lp := NewlookupPeer(addr, func(lp *lookupPeer){
            ci := make(map[string]interface{})
            ci["hostname"] = hostname
            ci["tcp_address"] = s.Opts.TCPAddress
            ci["http_address"] = s.Opts.HttpAddress
            ci["version"] = version.ServerVersion

            cmd, err := common.Identify(ci)
            if err != nil {
                lp.Close()
                return
            }
            resp, err := lp.Command(cmd)
            if err != nil {
                s.logf("lookup(%s): error %s - %s", lp, cmd, err)
            } else if bytes.Equal(resp, []byte("E_INVALID")) {
                s.logf("lookup(%s): lookup return %s", resp)
            } else {
                err := json.Unmarshal(resp, lp.Info)
                if err != nil {
                    s.logf("lookup(%s): error parse lookup-server response")
                } else {
                    s.logf("lookup(%s): lookup-server info: %+v", lp.Info)
                }
            }

            go func() {
                syncLocalFileChan <- lp
            }()
        })

        lp.Command(nil)
        s.lookupPeers = append(s.lookupPeers, lp)
    }

    ticker := time.Tick(15 * time.Second)
    loadTicker := time.Tick(60 * time.Second)
    for {
        select {
        case <- ticker:
            for _, lp := range(s.lookupPeers) {
                cmd := common.Ping()
                _, err := lp.Command(cmd)
                if err != nil {
                    s.logf("lookupd(%s): error %s - %s", lp, cmd, err)
                }
            }
        case lp := <- syncLocalFileChan:
            files := s.LocalFiles()
            cmd := common.Register(files)
            _, err := lp.Command(cmd)
            if err != nil {
                s.logf("lookupd(%s) error: %s - %s", lp, cmd, err)
            }

        case <- loadTicker:
            var pload *peerLoad
            var cmd *common.Command

            pload = s.getLoad()
            ploadBuf, err := json.Marshal(pload)
            if err != nil {
                s.logf("lookup(d) error: json.Marshal failed, err: %s", err)
                break
            }
            cmd = common.Load(ploadBuf)
            for _, lp := range(s.lookupPeers) {
                _, err := lp.Command(cmd)
                if err != nil {
                    s.logf("lookupd(%s) error: %s - %s", lp, cmd, err)
                }
            }
        case val := <- s.notifyChan:
            var cmd *common.Command
            var typ int32
            var files []string

            switch val.(type) {
            case *FileChange:
                typ = val.(*FileChange).typ
                files = []string{val.(*FileChange).file}

                switch (typ) {
                case ADD:
                    cmd = common.Register(files)
                case DEL:
                    cmd = common.UnRegister(files)
                }
            }
            for _, lp := range(s.lookupPeers) {
                _, err := lp.Command(cmd)
                if err != nil {
                    s.logf("lookupd(%s) error: %s - %s", lp, cmd, err)
                }
            }
        case <- s.exitChan:
            goto exit
        }
    }

exit:
    close(syncLocalFileChan)
    s.logf("lookupLoop quit")
}
