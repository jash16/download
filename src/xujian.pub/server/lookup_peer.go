package server

import (
    _"fmt"
    "net"
    "time"
    "xujian.pub/proto"
    "xujian.pub/common"
)

const (
    StateInit = iota
    stateDisconnected
    stateConnected
    StateSubscribed
    StateClosing
)

//lookupPeer do the job to connect the lookup server
type lookupPeer struct {
    ctx *context
    lookupSrvAddr string
    conn net.Conn
    state int32
    connectCallback func(*lookupPeer)
    Info peerInfo
}

type peerInfo struct {
    TCPPort          string    `json:"tcp_address"`
    //HTTPPort         int    `json:"http_address"`
    Hostname         string    `json:"hostname"`
    //MemTotal         int64  `json:"mem_total"`
    //CpuTotal         int64  `json:"cpu_total"`
    //Version          string `json:"version"`
    //BroadcastAddress string `json:"broadcast_address"`
}

type peerLoad struct {
    ClientNum  int64 `json:"client_num"`
    CpuUsage   float64 `json:"cpu_usage"`
    MemUsage   float64 `json:"mem_usage"`
}

func NewlookupPeer(addr string, callback func(*lookupPeer)) *lookupPeer {
    return &lookupPeer{
        lookupSrvAddr: addr,
        connectCallback: callback,
        state: stateDisconnected,
    }
}

func (l *lookupPeer) Connect() error {
    conn, err := net.DialTimeout("tcp", l.lookupSrvAddr, 5 * time.Second)
    if err != nil {
        l.ctx.s.logf("connect to lookup server(%s) failed: %s", l.lookupSrvAddr, err.Error())
        return err
    }
    l.conn = conn
    return nil
}

func (l *lookupPeer) String() string {
    return l.lookupSrvAddr
}

func (l *lookupPeer) Read(data []byte) (int, error) {
    l.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
    return l.conn.Read(data)
}

func (l *lookupPeer) Write(data []byte) (int, error) {
    l.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
    return l.conn.Write(data)
}

func (l *lookupPeer) Close() error {
    l.state = stateDisconnected
    return l.conn.Close()
}

func (lp *lookupPeer) Command(cmd *common.Command) ([]byte, error) {
    state := lp.state
    //init or reconnect
    if lp.state != stateConnected {
        err := lp.Connect()
        if err != nil {
            return nil, err
        }
        lp.state = stateConnected
        lp.Write([]byte("  V1"))
        if state == stateDisconnected {
            lp.connectCallback(lp)
        }
    }

    if (cmd == nil) {
        return nil, nil
    }

    _, err := cmd.WriteTo(lp)
    if err != nil {
        lp.Close()
        return nil, err
    }

    data, err := proto.ReadResponse(lp)
    if err != nil {
        lp.Close()
        return nil, err
    }
    return data, nil
}
