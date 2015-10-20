package lookup

import (
    "net"
    "bufio"
    "strings"
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

func (p *LookupProtocolV1)Exec(client *Client, reader *bufio.Reader, params [][]byte) ([]byte, error){
    cmd := string(param[0])
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
        return p.Ping(client, params[1:])
    }
    return nil, proto.NewFatalClientErr(nil, "E_INVALID", fmt.Sprintf("invalid command"))
}
