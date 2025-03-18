package wrapper

import (
	"context"
	"github.com/IndexStorm/common-go/termination"
)

type wrappedShutdowner struct {
	close func()
}

func NewShutdowner(close func()) termination.Shutdowner {
	return &Closer{close: close}
}

func (c *Closer) Shutdown(ctx context.Context) error {
	c.close()
	return nil
}
