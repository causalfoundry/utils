package util

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var (
	ErrNotBytes = errors.New("error not bytes")
)

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"message"`
	Data any    `json:"data"`
}

func NewErr(code int, msg string, data any) Err {
	return Err{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

func (e Err) Error() string {
	return e.Msg
}

func Panic(err error) {
	if err == nil {
		return
	}
	panic(err.Error())
}

func ErrNotFound(msg string) Err {
	return NewErr(http.StatusNotFound, msg, nil)
}

func OkJSON(ctx echo.Context, data any) error {
	return ctx.JSON(http.StatusOK, data)
}

func ErrBadRequest(msg string) error {
	return NewErr(http.StatusBadRequest, msg, nil)
}

func ErrInternal(msg string) error {
	return NewErr(http.StatusInternalServerError, msg, nil)
}

func ErrUnauthorized(msg string) error {
	return NewErr(http.StatusUnauthorized, msg, nil)
}

func ErrForbidden(msg string) error {
	return NewErr(http.StatusForbidden, msg, nil)
}

func Errs(err ...error) error {
	for _, e := range err {
		if e != nil {
			return e
		}
	}
	return nil
}
