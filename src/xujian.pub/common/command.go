package common

type Command struct {
    Name []byte
    Param [][]byte
    Body []byte
}

func (c *Command) WriteTo(w io.Writer) (int, error) {

}

func Ping() *Command {
    return &Command {
        Name: []byte("ping"),
    }
}

func Identify()
