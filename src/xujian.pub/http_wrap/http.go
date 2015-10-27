package http_wrap

import (
    "net"
    "fmt"
    "net/http"
    "xujian.pub/common"
)

func ServeHttp(listener net.Listener, handler http.Handler, l common.Logger) {
    l.Output(2, fmt.Sprintf("HTTP: listening on: %s", listener.Addr()))
    s := &http.Server {
        Handler: handler,
    }

    err := s.Serve(listener)
    if err != nil {
        l.Output(2, fmt.Sprintf("HTTP: error %s ", err))
    }
}
