package models

import "errors"

type ApiError struct {
	Error string
}

func NewApiError(msg string) ApiError {
	return ApiError{msg}
}

var (
	ErrMissingLessonData       = errors.New("lesson data is required for type 'lesson'")
	ErrMissingFreezeData       = errors.New("freeze data is required for type 'freeze'")
	ErrMissingProlongationData = errors.New("prolongation data is required for type 'prolongation'")
)