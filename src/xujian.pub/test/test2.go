package main

import (
    "encoding/json"
    "fmt"
)

type AddrStr struct {
    Addr string `json:"addr"`
}

func main() {
    ci := make(map[string]interface{})
    ci["addr"] = "0.0.0.0:1234"

    data, err := json.Marshal(ci)
    if err != nil {
        fmt.Printf("json marshal error: %s", err)
    }
    addr := AddrStr{}

    err = json.Unmarshal(data, &addr)
    if err != nil {
        fmt.Printf("json marshal error: %s", err)
    }
    fmt.Printf("data: %#v\n", addr)
}
