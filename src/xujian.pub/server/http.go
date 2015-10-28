package server

import (
    "io"
    "fmt"
    "strings"
    "net/url"
    "io/ioutil"
    "net/http"
)

type httpServer struct {
    ctx *context
}

func (h *httpServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
    switch(r.URL.Path) {
    case "/ping":
        h.pingHandler(rw, r)
    case "/lookup":
        h.lookupHandler(rw, r)
    case "/download":
        h.downHandler(rw, r)
    }
}

func (h *httpServer) pingHandler(rw http.ResponseWriter, r *http.Request) {
    rw.Header().Set("Content-length", "2")
    io.WriteString(rw, "OK")
}

func (h *httpServer) lookupHandler(rw http.ResponseWriter, r *http.Request) {
    files := h.ctx.s.LocalFiles()
    file_str := strings.Join(files, " ")
    length := len(file_str)
    rw.Header().Set("Content-length", fmt.Sprintf("%d", length))
    io.WriteString(rw, file_str)
}

func (h *httpServer) downHandler(rw http.ResponseWriter, r *http.Request) {
    var err error
    var data []byte
    var statusCode int
    var file string
    var fullFile string
    values, err := url.ParseQuery(r.URL.RawQuery)
    if err != nil {
        h.ctx.s.logf("download error: %s", err)
        data = []byte("internal error")
        statusCode = 500
        goto end
    }
    file = values.Get("file")
    fullFile = h.ctx.s.Opts.DataPath + "/" + file
    data, err = ioutil.ReadFile(fullFile)
    if err != nil {
        h.ctx.s.logf("download error: %s", err)
        if strings.Contains(err.Error(), "no such file or directory") {
            statusCode = 404
            data = []byte("not found")
        } else {
            data = []byte("internal error")
            statusCode = 500
        }
        goto end
    }
    statusCode = 200
end:
    length := len(data)
    rw.Header().Set("Content-Length", fmt.Sprintf("%d", length))
    rw.WriteHeader(statusCode)
    //rw.Header().Set("Status-code", fmt.Sprintf("%d", statusCode))
    io.WriteString(rw, string(data))
}
