package conf

import (
    "fmt"
)

type ConfError struct {
    message string
    filename string
    lineno int
}

func (e *ConfError) Error() string {
    return fmt.Sprintf("%s in '%s' line %d", e.message, e.filename, e.lineno)
}
