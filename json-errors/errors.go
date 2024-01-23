package json_errors

import "errors"

var (
	ErrDuplicateFieldInJSON = errors.New("duplicate field found in provided JSON")
)
