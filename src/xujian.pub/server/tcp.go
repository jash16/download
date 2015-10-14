package server

import (
    "io"
    "net"
    "xujian.pub/proto"
)

type tcpServer struct {
    ctx *context
}

func (t *tcpServer) Handle(clientConn net.Conn) {
    t.ctx.s.logf("MyServer: new client(%s)", clientConn.RemoteAddr())

    buf := make([]byte, 4)
    _, err := io.ReadFull(clientConn, buf)
    if err != nil {
        t.ctx.s.logf("ERROR: failed to read client(%s) protocol version - %s", 
                      clientConn.RemoteAddr(), err)
        return
    }
    protocolMagic := string(buf)
    var prot proto.Protocol
    t.ctx.s.logf("MyServer: %s", protocolMagic)
    switch protocolMagic {
    case "  V1": //when upgrade, V2, V3..., and can also support multi-version
        t.ctx.s.logf("MyServer: magic v1")
        prot = &protocolV1{ctx: t.ctx}
    default:
        proto.SendFrameResponse(clientConn, 3, []byte("E_BAD_PROTOCOL"))
        clientConn.Close()
        t.ctx.s.logf("ERROR: client(%s) bad version")
    }

    err = prot.IOLoop(clientConn)
    if err != nil {
        t.ctx.s.logf("ERROR: client(%s) - %s", clientConn.RemoteAddr(), err)
        return
    }
}

