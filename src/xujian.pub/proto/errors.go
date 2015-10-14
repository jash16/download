package proto

type ChildErr interface {
    Parent() error
}

type ClientErr struct {
    ParentErr error
    Code string
    Desc string
}

func (c *ClientErr) Error() string {
    return c.Code + " " + c.Desc
}

func (c *ClientErr) Parent() error {
    return c.ParentErr
}

func NewClientErr(parent error, code string, desc string) *ClientErr {
    return &ClientErr{
        ParentErr: parent,
        Code: code,
        Desc: desc,
    }
}

type FatalClientErr struct {
    ParentErr error
    Code string
    Desc string
}

func (f *FatalClientErr) Error() string {
    return f.Code + " " + f.Desc
}

func (f *FatalClientErr) Parent() error {
    return f.ParentErr
}

func NewFatalClientErr(parent error, code, desc string) *FatalClientErr {
    return &FatalClientErr {
        ParentErr: parent,
        Code: code,
        Desc: desc,
    }
}
