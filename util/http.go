package util

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type StrMap map[string]string

type CustomValidator struct {
	Validator *validator.Validate
}

var Validator = validator.New()

func Bind(ctx echo.Context, a any) error {
	if err := ctx.Bind(a); err != nil {
		return err
	}

	err := Validator.Struct(a)
	if err != nil {
		return ErrBadRequest(err.Error())
	}
	return nil
}

func RespNoContent(ctx echo.Context, err error) error {
	switch err {
	case nil:
		return ctx.NoContent(http.StatusOK)
	default:
		return errHandle(ctx, err)
	}
}

func RespCSV(ctx echo.Context, err error, csv string) error {
	ctx.Response().Header().Set("Content-Type", "text/csv")
	ctx.Response().Header().Set("Content-Disposition", "attachment; filename=\"persons.csv\"")

	switch err {
	case nil:
		return ctx.String(http.StatusOK, csv)
	default:
		return errHandle(ctx, err)
	}
}

func RespJSON(ctx echo.Context, err error, payload any) error {
	switch err {
	case nil:
		return ctx.JSON(http.StatusOK, payload)
	default:
		return errHandle(ctx, err)
	}
}

func errHandle(ctx echo.Context, err error) error {
	if e, ok := err.(Err); ok {
		return ctx.JSON(e.Code, e)
	}
	switch err {
	case sql.ErrNoRows:
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	default:
		return ErrInternal(err.Error())
	}
}

func PostJSON[REQ, RESP any](req REQ, url string) (ret RESP, err error) {
	if err = Validator.Struct(req); err != nil {
		return
	}

	bs, err := json.Marshal(req)
	if err != nil {
		return
	}

	resp, e := http.Post(url, "application/json", bytes.NewBuffer(bs))
	if e != nil {
		err = fmt.Errorf("error post request: %w", e)
		return
	}

	b, e := io.ReadAll(resp.Body)
	if e != nil {
		err = fmt.Errorf("error read body: %w, body: %s", e, string(b))
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("status code not 200: %s", string(b))
		return
	}

	if e = json.Unmarshal(b, &ret); e != nil {
		err = fmt.Errorf("error in unmarshalling response, err: %w, response: %s", e, string(b))
		return
	}

	if e = Validator.Struct(ret); e != nil {
		err = fmt.Errorf("response not valid, err: %w, response: %+v", e, ret)
		return
	}

	return
}

// ToMid
// means ToMIddleware
// convert HandlerFunc to one type of middleware, next() is at the bottom
func ToMid(handler echo.HandlerFunc) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			if err := handler(ctx); err != nil {
				return err
			}
			return next(ctx)
		}
	}
}

func CheckHeaderEqual(ctx echo.Context, header, value string) bool {
	return ctx.Request().Header.Get(header) == value
}

func OkJSON(ctx echo.Context, data any) error {
	return ctx.JSON(http.StatusOK, data)
}

func ErrBadRequest(msg string) error {
	return echo.NewHTTPError(http.StatusBadRequest, msg)
}

func ErrInternal(msg string) error {
	return echo.NewHTTPError(http.StatusInternalServerError, msg)
}

func ErrUnauthorized(msg string) error {
	return echo.NewHTTPError(http.StatusUnauthorized, msg)
}

func ErrForbidden(msg string) error {
	return echo.NewHTTPError(http.StatusForbidden, msg)
}

func GetPayloads[T any](ctx echo.Context) (t []T, err error) {
	payload := ctx.Get("payload")

	switch payload {
	case nil:
		if err = ctx.Bind(&t); err != nil {
			err = fmt.Errorf("failed to bind payload slice of %T: %w", *new(T), err)
			return
		}
		ctx.Set("payload", t)
		payload = t
	}

	t, ok := payload.([]T)
	if !ok {
		err = ErrBadRequest("unable to get payload")
	}

	for i := range t {
		if err = Validator.Struct(t[i]); err != nil {
			err = echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
	}

	return
}

func GetQuery[T int | string](ctx echo.Context, name string) (t T, err error) {
	switch any(t).(type) {
	case int:
		tt, err := strconv.Atoi(ctx.QueryParam(name))
		if err != nil {
			err = ErrBadRequest(err.Error())
			return t, err
		}
		t = any(tt).(T)
	case string:
		t = any(ctx.QueryParam(name)).(T)
	}
	return
}

func GetParam[T int | string](ctx echo.Context, name string) (t T, err error) {
	switch any(t).(type) {
	case int:
		tt, err := strconv.Atoi(ctx.Param(name))
		if err != nil {
			err = ErrBadRequest(err.Error())
			return t, err
		}
		t = any(tt).(T)
	case string:
		t = any(ctx.Param(name)).(T)
	}
	return
}

func GetID[T int | string](ctx echo.Context) (t T, err error) {
	return GetParam[T](ctx, "id")
}

// GetPayload takes care of marshaling the payload from request
// 1) if the payload was never read before, then it read the payload, and it saves within the context
// 2) if the payload has already been read, then it read directly from the context
// can only work if the http verb is NOT [GET DELETE HEAD] for payload in body
func GetPayload[T any](ctx echo.Context) (t T, err error) {
	name := fmt.Sprintf("%T", t)
	payload := ctx.Get(name)

	switch payload {
	case nil:
		if err = ctx.Bind(&t); err != nil {
			err = ErrBadRequest(err.Error())
			return
		}

		switch ctx.QueryParam("no_check") {
		case "true":
		default:
			if err = Validator.Struct(t); err != nil {
				b, _ := json.Marshal(t)
				err = ErrBadRequest(fmt.Sprintf("err: %s, payload: %s", err.Error(), string(b)))
				return
			}

			if c, ok := any(t).(Checkable); ok {
				if err = c.Check(); err != nil {
					err = ErrBadRequest(fmt.Sprintf("checkable fail: %s", err.Error()))
					return
				}
			}
		}

		ctx.Set(name, t)
		return
	}

	t, ok := payload.(T)
	if !ok {
		err = ErrBadRequest("unable to get payload")
		return
	}

	return
}

type ReqCtx struct {
	Req  *http.Request
	Resp *httptest.ResponseRecorder
	Ctx  echo.Context
}

type RequestCfg struct {
	Method      string
	URL         string
	Headers     StrMap
	QueryParams StrMap
	Cookies     []http.Cookie
	Payload     any // should be a struct
	CtxKVs      Obj
}

func NewHttpTestKit(engine *echo.Echo, r RequestCfg) ReqCtx {
	objBytes, err := json.Marshal(r.Payload)
	Panic(err)

	resp := httptest.NewRecorder()
	if r.URL == "" {
		r.URL = "http://example.com/"
	}
	req := httptest.NewRequest(r.Method, r.URL, bytes.NewReader(objBytes))

	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	for _, c := range r.Cookies {
		req.AddCookie(&c)
	}

	q := req.URL.Query()
	for k, v := range r.QueryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	ctx := engine.NewContext(req, resp)
	for k, v := range r.CtxKVs {
		ctx.Set(k, v)
	}

	return ReqCtx{
		Req:  req,
		Resp: resp,
		Ctx:  ctx,
	}
}

func ParamInt(ctx echo.Context, name string) (int, error) {
	return strconv.Atoi(ctx.Param(name))
}