package client

type BackendWork interface {
    Process(start, end int64, file string, data []byte) error
}
