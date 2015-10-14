package client

type BackendWork interface {
    Put(start int64, data []byte) error
}
