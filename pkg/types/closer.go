package types

import "context"

type Closer interface {
	Close(ctx context.Context) error
}
