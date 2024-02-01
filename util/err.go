package util

import "errors"

var (
	ErrNotBytes                        = errors.New("error not bytes")
	ErrCannotDeleteModelWithInvPurpose = errors.New("cannot delete model with inv purpose")
)

type Err struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
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
