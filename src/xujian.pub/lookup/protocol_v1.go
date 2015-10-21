package lookup

import (
    "os"
    "net"
    "fmt"
    "json"
    "bufio"
    "sync/atomic"
    "strings"
    "xujian.pub/proto"
)

type LookupProtocolV1 struct {
    ctx *context
}

func (l *LookupProtocolV1)IOLoop(conn net.Conn) error {
    var err error
    var line string

    err = nil
    client := NewClientV1(conn)
    reader := bufio.NewReader(conn)
    for {
        line, err := reader.ReadString("\n")
        if err != nil {
            break
        }

        line = strings.TrimSpace(line)
        params := strings.Split(line, " ")

        resp, err := p.Exec(client, reader, params)
        if err != nil {
            p.ctx.s.logf("client(%s) exec error- %s", client, err)
            break
        }

        if resp != nil {
            _, err := proto.SendResponse(client, resp)
            if err != nil {
                p.ctx.s.logf("send response to client(%) error - %s", client, err)
                break
            }
        }
    }
    p.ctx.s.logf("closing client(%s)", client)
    //client may download client or server
    if client.peerInfo != nil {
        
    }
    client.Close()
    return err
}

func (p *LookupProtocolV1)Exec(client *Client, reader *bufio.Reader, params []string) ([]byte, error){
    cmd := param[0]
    switch(cmd) {
    case "IDENTIFY":
        return p.Identify(client, reader, params[1:])
    case "REGISTER":
        return p.Register(client, reader, params[1:])
    case "UNREGISTER":
        return p.UnRegister(client, reader, params[1:])
    case "LOOKUP":
        return p.Lookup(client, reader, params[1:])
    case "LOOKUPALL":
        return p.LookupAll(client, reader, params[1:])
    case "LOAD":
        return p.Load(client, reader, params[1:])
    case "PING":
        return p.Ping(client)
    }
    return nil, proto.NewFatalClientErr(nil, "E_INVALID", fmt.Sprintf("invalid command"))
}

func (p *LookupProtocolV1) Ping(client *ClientV1) ([]byte, error){
    if client.peerInfo != nil {
        cur := time.Now().Unix()
        last := time.Unix(atomic.LoadInt64(&client.peerInfo.lastActive), 0)
        p.ctx.s.logf("client(%s) pinged last(%s)", client, last)
        atomic.StoreInt64(&client.peerInfo.lastActive, cur)
    }
    return []byte("OK"), nil
}

func (p *LookupProtocolV1) Identify(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error){
    var buf []byte
    var err error
    //buf := make([]buf, 4)

    if client.peerInfo != nil {
        return nil, proto.NewFatalClientErr(nil, "E_INVALID", fmt.Sprintf("client cannot identify again"))
    }

    var bodyLen int32
    err = binary.Read(reader, binary.BigEndian, &bodyLen)
    if err != nil {
        return nil, proto.NewFataClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }

    buf := make([]byte, bodyLen)
    n, err := io.ReadFull(reader, buf)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }

    peerInfo := PeerInfo {id: client.RemoteAddr().String()}
    err := json.Unmarshal(buf, &peerInfo)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }

    peerInfo.RemoteAddr = client.RemoteAddr().String()

    tnow := time.Now().Unix()
    atomic.StoreInt64(&peerInfo.lastActive)

    client.peerInfo = &peerInfo

    data := make(map[string]interface{})
    hostname, err := os.Hostname()
    if err != nil {
        p.ctx.s.logf("FATAL, quit now")
        os.Exit(1)
    }
    data["hostname"] = hostname

    buf, err := json.Marshal(data)
    if err != nil {
        return []byte("OK"), nil
    }
    return buf, nil
}
