package service

import "errors"

var (
	ErrNotFound     = errors.New("service: not found")
	ErrInvalidInput = errors.New("service: invalid input")
)
