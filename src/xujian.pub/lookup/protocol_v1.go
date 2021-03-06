package lookup

import (
    "os"
    "net"
    "io"
    "time"
    "fmt"
    "encoding/binary"
    "encoding/json"
    "bufio"
    "sync/atomic"
    "strings"
    "xujian.pub/proto"
)

type LookupProtocolV1 struct {
    ctx *context
}

func (p *LookupProtocolV1)IOLoop(conn net.Conn) error {
    var err error
    //var line string

    err = nil
    client := NewClientV1(conn)
    reader := bufio.NewReader(conn)
    for {
        line, err := reader.ReadString('\n')
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
        p.ctx.s.Hold.RemoveProducer(client.peerInfo)
    }
    client.Close()
    return err
}

func (p *LookupProtocolV1)Exec(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error){
    cmd := params[0]
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
        return nil, proto.NewFatalClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }

    buf = make([]byte, bodyLen)
    _, err = io.ReadFull(reader, buf)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }

    peerInfo := PeerInfo {id: client.RemoteAddr().String()}
    err = json.Unmarshal(buf, &peerInfo)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_BAD_BODY", fmt.Sprintf("IDENTIFY failed"))
    }
    p.ctx.s.logf("register: %+v", peerInfo)
    //peerInfo.RemoteAddr = client.RemoteAddr().String()

    tnow := time.Now().Unix()
    atomic.StoreInt64(&peerInfo.lastActive, tnow)

    client.peerInfo = &peerInfo

    data := make(map[string]interface{})
    hostname, err := os.Hostname()
    if err != nil {
        p.ctx.s.logf("FATAL, quit now")
        os.Exit(1)
    }
    data["hostname"] = hostname
    data["tcp_address"] = p.ctx.s.Opts.TcpAddress
    buf, err = json.Marshal(data)
    if err != nil {
        return []byte("OK"), nil
    }
    p.ctx.s.logf("client: %s identify success", client)
    return buf, nil
}

func (p *LookupProtocolV1) Register(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error){
    var buf []byte
    var bodyLen int32

    err := binary.Read(reader, binary.BigEndian, &bodyLen)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("REGISTER failed"))
    }
    buf = make([]byte, bodyLen)
    _, err = io.ReadFull(reader, buf)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("REGISTER failed"))
    }

    files := strings.Split(string(buf), "+")
    for _, file := range files {
        p.ctx.s.Hold.AddProducer(file, client.peerInfo)
    }
    p.ctx.s.logf("client %s REGISTER %s", client, buf)
    return []byte("OK"), nil
}

func (p *LookupProtocolV1) UnRegister(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error) {
    var buf []byte
    var bodyLen int32
    var err error

    err = binary.Read(reader, binary.BigEndian, &bodyLen)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("UNREGISTER failed"))
    }

    buf = make([]byte, bodyLen)
    _, err = io.ReadFull(reader, buf)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("UNREGISTER failed"))
    }

    files := strings.Split(string(buf), "+")

    for _, file := range(files) {
        p.ctx.s.Hold.RemoveFileProducer(file, client.peerInfo)
    }

    p.ctx.s.logf("client %s UNREGISTER file: %s", client, buf)
    return []byte("OK"), nil
}

func (p *LookupProtocolV1) Load(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error){
    var buf []byte
    var bodyLen int32
    var err error

    err = binary.Read(reader, binary.BigEndian, &bodyLen)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("LOAD failed"))
    }

    buf = make([]byte, bodyLen)
    _, err = io.ReadFull(reader, buf)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INVALID_BODY", fmt.Sprintf("LOAD failed"))
    }

    p.ctx.s.logf("client: %s, load: %s", client, buf)
    //client.peerInfo.UpdateLoad()
    return []byte("OK"), nil
}

func (p *LookupProtocolV1) Lookup(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error){
    file := params[0]

    producer := p.ctx.s.Hold.FindProperProducer(file)
    if producer == nil {
        return []byte("E_NOT_FOUND"), nil
    }

    pbuf, err := json.Marshal(producer.peerInfo)
    if err != nil {
        return nil, proto.NewFatalClientErr(err, "E_INTERNAL_ERROR", fmt.Sprintf("LOOKUP failed"))
    }
    p.ctx.s.logf("send lookup data: %s", pbuf)
    return pbuf, nil
}

func (p *LookupProtocolV1) LookupAll(client *ClientV1, reader *bufio.Reader, params []string) ([]byte, error) {
    file := params[0]
    producers := p.ctx.s.Hold.FindProducers(file)
    if producers == nil {
        return []byte("E_NOT_FOUND"), nil
    }

    data, err := json.Marshal(struct {
        File string
        Data interface{}
    }{
        File: file,
        Data: producers.PeerInfo(),
    })
    if err != nil {
        return []byte("E_NOT_FOUND"), nil
    }
    return data, nil
}
