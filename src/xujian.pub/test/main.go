package main

import (
    "log"
    "fmt"
    "strings"
    "github.com/howeyc/fsnotify"
)

type Cache map[string]string

var ca Cache

func main() {
    ca = make(map[string]string)
    ca["hello"] = "world"
    fmt.Printf("%#v\n", ca)
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        log.Fatal(err)
    }

    done := make(chan bool)
    // Process events
    go func() {
    for {
        select {
        case ev := <-watcher.Event:
            paths := strings.Split(ev.Name, "/")
            filename := paths[len(paths) - 1]
            log.Println("file: %s", filename)
            log.Println("event:", ev)
        case err := <-watcher.Error:
            log.Println("error:", err)
        }
    }
    }()

    err = watcher.Watch("/tmp")
    if err != nil {
        log.Fatal(err)
    }

    // Hang so program doesn't exit
    <-done
    watcher.Close()
}
