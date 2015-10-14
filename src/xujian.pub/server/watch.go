package server

import (
    "os"
    "strings"
    "github.com/howeyc/fsnotify"
)
func (s *Server)watchLoop() {
    s.logf("file watch start")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		s.logf(err.Error())
        os.Exit(1)
	}

    done := make(chan bool)
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
                s.processEvent(ev)
			case err := <-watcher.Error:
				s.logf("watch error:", err)
			}
		}
	}()

	err = watcher.Watch(s.Opts.DataPath)
	if err != nil {
	    s.logf(err.Error())
        os.Exit(1)
	}
    <-done
    watcher.Close()
    s.logf("file watch quit")
}

func (s *Server)processEvent(ev *fsnotify.FileEvent) {
    var typ int
    if ev.IsCreate() {
        typ = ADD
    }else if ev.IsDelete() {
        typ = DEL
    }else if ev.IsModify() {
        typ = MOD
    }else if ev.IsRename() {
        typ = RENAME
    }else if ev.IsAttrib() {
        typ = MOD
    }
    paths := strings.Split(ev.Name, "/")
    filename := paths[len(paths) - 1]
    cev := &cacheEvent {
        eventType: typ,
        filename: filename,
    }
    s.cacheEventChan <- cev
}
