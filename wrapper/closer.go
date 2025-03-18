package wrapper

import "io"

type Closer struct {
	close func()
}

func NewCloser(close func()) io.Closer {
	return &Closer{close: close}
}

func (c *Closer) Close() error {
	c.close()
	return nil
}
