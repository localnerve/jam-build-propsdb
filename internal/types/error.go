package types

import "fmt"

type CustomError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%d: %s [type: %s]", e.Code, e.Message, e.Type)
}
