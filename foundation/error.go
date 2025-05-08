package foundation

import (
	"fmt"

	"github.com/pkg/errors"
)

func ErrWithStack(message string, args ...any) error {
	return errors.WithStack(fmt.Errorf(message, args...))
}
