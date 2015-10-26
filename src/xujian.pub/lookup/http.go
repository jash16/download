package lookup

import (
    "net/http"
)

type httpServer struct {
    ctx *conext
}

func (*h httpServer) ServeHttp(rw http.ResponseWriter, r *http.Request) {

}
