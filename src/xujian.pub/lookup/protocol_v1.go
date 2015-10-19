package lookup

import (
    "net"
)

type LookupProtocolV1 struct {
    ctx *context
}

func (l *LookupProtocolV1)IOLoop(conn net.Conn) error {
    client := NewClientV1(conn)

    for {
        
    }

    return nil
}


