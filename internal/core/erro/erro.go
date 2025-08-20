package erro

import (
	"errors"
)

var (
	ErrNotFound 		= errors.New("item not found")
	ErrTimeout			= errors.New("timeout: context deadline exceeded.")
)