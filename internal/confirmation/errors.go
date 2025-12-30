package confirmation

import "errors"

var (
	// Question validation errors
	ErrEmptyQuestion       = errors.New("question cannot be empty")
	ErrHeaderTooLong       = errors.New("header must be max 12 characters")
	ErrInvalidOptionCount  = errors.New("must have 2-4 options")
	ErrEmptyOptionLabel    = errors.New("option label cannot be empty")
	ErrInvalidQuestionCount = errors.New("must have 1-4 questions")

	// Answer validation errors
	ErrNoSelection         = errors.New("no option selected")
	ErrInvalidSelection    = errors.New("invalid option selected")
)
