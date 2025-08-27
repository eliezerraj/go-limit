package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrBadRequest 		= errors.New("bad request ! check parameters")
	ErrTimeout			= errors.New("timeout: context deadline exceeded.")
)