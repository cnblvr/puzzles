package app

import (
	"fmt"
)

//go:generate stringer -type=statusCode -linecomment -trimprefix Status -output errors_string.go

const (
	StatusBadRequest          = statusCode(400)  // Bad request
	StatusUnauthorized        = statusCode(401)  // Unauthorized
	StatusMethodNotAllowed    = statusCode(405)  // Method not allowed
	StatusInternalServerError = statusCode(500)  // Internal server error
	StatusUnknown             = statusCode(9999) // Unknown error
)

//var (
//	StatusBadRequest          = Status(&status{code: 400, msg: "Bad request"})
//	StatusUnauthorized        = Status(&status{code: 401, msg: "Unauthorized"})
//	StatusMethodNotAllowed    = Status(&status{code: 405, msg: "Method not allowed"})
//	StatusInternalServerError = Status(&status{code: 500, msg: "Internal server error"})
//	StatusUnknown             = Status(&status{code: 9999, msg: "Unknown error"})
//)

type Status interface {
	WithMessage(msg string) Status
	WithError(err error) Status
	Error() string
	GetError() error
	GetMessage() string
	GetCode() uint16
}

type status struct {
	code statusCode
	msg  string
	err  error
}

func (s status) WithMessage(msg string) Status {
	s.msg = msg
	return &s
}

func (s status) WithError(err error) Status {
	s.err = err
	return &s
}

func (s status) Error() string {
	out := fmt.Sprintf("status %d", s.code)
	if s.msg != "" {
		out += fmt.Sprintf(": %s", s.msg)
	} else {

	}
	if s.err != nil {
		out += fmt.Sprintf("; %v", s.err)
	}
	return out
}

func (s status) GetError() error {
	return s.err
}

func (s status) GetMessage() string {
	return s.msg
}

func (s status) GetCode() uint16 {
	return uint16(s.code)
}

type statusCode uint16

func (i statusCode) WithMessage(msg string) Status {
	return &status{
		code: i,
		msg:  msg,
	}
}

func (i statusCode) WithError(err error) Status {
	return &status{
		code: i,
		err:  err,
	}
}

func (i statusCode) Error() string {
	return fmt.Sprintf("status %d: %s", i, i.String())
}

func (i statusCode) GetError() error {
	return nil
}

func (i statusCode) GetMessage() string {
	return ""
}

func (i statusCode) GetCode() uint16 {
	return uint16(i)
}
