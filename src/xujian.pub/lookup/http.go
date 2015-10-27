package lookup

import (
    "io"
    "fmt"
    "net/url"
    "encoding/json"
    "net/http"
)

type httpServer struct {
    ctx *context
}

func (h *httpServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    switch r.URL.Path {
    case "/ping":
        h.pingHandler(rw, r)
    case "/lookup":
        h.lookupHandler(rw, r)
    }
}

func (h *httpServer) pingHandler(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-length", "2")
    io.WriteString(rw, "OK")
}

func (h *httpServer) lookupHandler(rw http.ResponseWriter, r *http.Request) {
    var data []byte
    var err error
    h.ctx.s.logf("query string: %s", r.URL.RawQuery)
    values,_ := url.ParseQuery(r.URL.RawQuery)
    if values == nil {
        h.ctx.s.logf("values is nil")
    }
    file := values.Get("file")

    producers := h.ctx.s.Hold.FindProducers(file)
    if producers == nil {
        //return []byte("E_NOT_FOUND"), nil
        h.ctx.s.logf("not found producer for file: %s", file)
        data = []byte("not found")
        //return
        goto END
    }

    data, err = json.Marshal(struct {
        File string
        Data interface{}
    }{
        File: file,
        Data: producers.PeerInfo(),
    })
    if err != nil {
        //h.ctx.s.logf("not found producer for file: %s", file)
        data = []byte("not found")
    }
END:
    length := len(data)
    rw.Header().Set("Content-length", fmt.Sprintf("%d", length))
    io.WriteString(rw, string(data))
    //h.ctx.s.logf("lookup: %s", file)
}
