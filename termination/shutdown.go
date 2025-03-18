package termination

import (
	"context"
	"io"
)

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type closerShutdowner struct {
	closer io.Closer
}

func ShutdownerFromCloser(closer io.Closer) Shutdowner {
	return &closerShutdowner{closer: closer}
}

func (s *closerShutdowner) Shutdown(ctx context.Context) error {
	return s.closer.Close()
}
