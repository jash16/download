package server

import (
    "os"
    "time"
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
            ci["tcp_address"] = n.Opts.TCPAddress
            ci["http_address"] = n.Opts.HttpAddress
            ci["version"] = common.ServerVersion

            cmd, err := common.Identify(ci)
            if err != nil {
                lp.Close()
                return
            }
            resp, err := lp.Command(cmd)
            if err != nil {
                n.logf("lookup(%s): error %s - %s", lp, cmd, err)
            } else if bytes.Equal(resp, []byte("E_INVALID")) {
                n.logf("lookup(%s): lookup return %s", resp)
            } else {
                err := json.Unmarshal(resp, lp.Info)
                if err != nil {
                    n.logf("lookup(%s): error parse lookup-server response")
                } else {
                    n.logf("lookup(%s): lookup-server info: %+v", lp.Info)
                }
            }

            go func() {
                syncLocalFileChan <- lp
            }()
        })

        lp.Command(nil)
        n.lookupPeers = append(n.lookupPeers, lp)
    }

    ticker := time.Tick(15 * time.Second)
    for {
        select {
        case <- ticker:
            for _, lp := range(s.lookupPeers) {
                cmd := common.Ping()
                _, err := lp.Command(cmd)
                if err != nil {
                    n.logf("lookupd(%s): error %s - %s", lp, cmd, err)
                }
            }
        case lp := <- syncLocalFileChan:
            files := s.LocalFiles()
            cmd := common.Register(files)
            _, err := lp.Command(cmd)
            if err != nil {
                s.logf("lookupd(%s) error: %s - %s", lp, cmd, err)
            }

        case val := <- s.notifyChan:
            switch(val.(type)) {
            case 
            }
        case <- s.exitChan:
            goto exit
        }
    }
exit:
    s.logf("lookupLoop quit")
}
