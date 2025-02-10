package errors

import "fmt"

type ChessError struct {
	Op  string
	Err error
}

func (e *ChessError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("%s failed", e.Op)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ChessError) Unwrap() error {
	return e.Err
}
