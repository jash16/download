package client

type BackendWork interface {
    Process(start int64, file string, data []byte) error
}
