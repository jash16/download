package client

import (
    "net"
    "time"
    "encoding/json"
    "xujian.pub/proto"
    "xujian.pub/common"
)

type PeerInfo struct {
    TcpAddr string `json:"tcp_address"`
    HttpAddr string `json:"http_address"`
    Hostname string `json:"hostname"`
    Version string `json:"version"`
}

func GetPeerInfo(file string, addr string) (*PeerInfo , error){
    peerInfo := &PeerInfo{}

    dialer := &net.Dialer {
        Timeout: 5 * time.Second,
    }

    conn, err := dialer.Dial("tcp", addr)
    if err != nil {
        return nil, err
    }
    _, err = conn.Write([]byte("  V1"))
    if err != nil {
        return nil, err
    }
    cmd := common.Lookup(file)
    _, err = cmd.WriteTo(conn)
    if err != nil {
        return nil, err
    }
    resp, err := proto.ReadResponse(conn)
    if err != nil {
        println("receive error  response")
        return nil, err
    }
    println(string(resp))
    err = json.Unmarshal(resp, peerInfo)
    if err != nil {
        return nil, err
    }
    return peerInfo, nil
}
