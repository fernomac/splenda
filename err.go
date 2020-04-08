package splenda

import "fmt"

// Error is a structured error from Splenda.
type Error struct {
	HTTP    int    `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v: %v", e.Code, e.Message)
}

var (
	// ErrInsufficientCoins is the error returned when the user tries to make a
	// move but there are not enough coins either in the bank or in their hand.
	ErrInsufficientCoins error = &Error{
		HTTP:    400,
		Code:    "InsufficientCoins",
		Message: "not enough coins available to do that",
	}
)
