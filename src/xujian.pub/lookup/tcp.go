package lookup

import (
    "net"
    "io"
 _   "xujian.pub/common"
    "xujian.pub/proto"
)

type tcpServer struct {
    ctx *context
}

func (p *tcpServer) Handle(conn net.Conn) {
    buf := make([]byte, 4)

    _, err := io.ReadFull(conn, buf)
    if err != nil {
        p.ctx.s.logf("error: failed to read protocol - %s", err)
        return
    }

    protoMagic := string(buf)
    p.ctx.s.logf("client(%s) send protocol: %s", conn.RemoteAddr(), protoMagic)
    var protocol proto.Protocol
    switch(protoMagic) {
    case "  V1":
        protocol = &LookupProtocolV1{ctx: p.ctx}
    default:
        proto.SendResponse(conn, []byte("E_BAD_PROTOCOL"))
        p.ctx.s.logf("client(%s) wrong protocol: %s", conn.RemoteAddr(), protoMagic)
        conn.Close()
        return
    }

    err = protocol.IOLoop(conn)
    if err != nil {
        p.ctx.s.logf("client(%s) error - %s", conn.RemoteAddr(), err)
        return
    }
}
