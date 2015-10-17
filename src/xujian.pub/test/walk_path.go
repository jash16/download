package main

import (
    "fmt"
    "os"
    "strings"
    "path/filepath"
)

func main() {
    root := "/home/xujian/files"
    if root[len(root) - 1] != '/' {
        root = fmt.Sprintf("%s/", root)
    }
    files := []string{"test"}
    filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if ! info.IsDir() {
            fmt.Printf("%s\n", path)
            fi := strings.TrimPrefix(path, root)
            files = append(files, fi)
        }
        return nil
    })

    for _, file := range(files) {
        fmt.Printf("%s\n", file)
    }
}
