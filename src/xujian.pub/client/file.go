package client

type fileInfo struct {
    filename string
    filemd5  string
    fileSize int64
    blocks   map[string]*blockInfo
}

type blockInfo struct {
     blockmd5 string
     blockStart int64
     blockEnd   int64
     blockSize int64
     status    int
}

func NewfileInfo(md5 string, fileSize int64, filename string) *fileInfo {
    return &fileInfo {
        filemd5: md5,
        fileSize: fileSize,
        filename: filename,
        blocks: make(map[string]*blockInfo),
    }
}

/*
func (f *fileInfo) AddBlock(b *blockInfo) {
    
}
*/
