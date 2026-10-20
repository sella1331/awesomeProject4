package errors

import "errors"

var (
	ErrTaskIsEmpty       = errors.New("not found")
	ErrTaskAlreadyExists = errors.New("task already exists")
	ErrNotFind           = errors.New("not find")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidCreds      = errors.New("invalid creds")
	ErrNotValidToken     = errors.New("not valid token")
	ErrNotValidPassword  = errors.New("not valid password")
)
