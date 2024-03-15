package util

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type StrMap map[string]string

var CustomErrHandler = func(e *echo.Echo) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		// Check if the error is your custom error type
		if e, ok := err.(Err); ok {
			// Write to context or modify response based on custom error
			e1 := c.JSON(e.Code, map[string]any{"message": e.Msg, "data": fmt.Sprintf("%+v", e.Data)})
			if e1 != nil {
				_ = c.JSON(http.StatusInternalServerError, "error marshal error json: "+e1.Error())
			}
			return
		}

		// Handle other errors or default to Echo's internal error handling
		e.DefaultHTTPErrorHandler(err, c)
	}
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

func RespCSV(ctx echo.Context, err error, csv, fileName string) error {
	ctx.Response().Header().Set("Content-Type", "text/csv")
	ctx.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))

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

func buildReq(method, url string, query, header map[string]string, data any, ctx context.Context) (ret *http.Request, err error) {
	var queryStr []string
	for k, v := range query {
		queryStr = append(queryStr, fmt.Sprintf("%s=%s", k, v))
	}
	if len(queryStr) != 0 {
		url = url + "?" + strings.Join(queryStr, "&")
	}

	b, _ := json.Marshal(data)
	if ctx == nil {
		ret, err = http.NewRequest(method, url, bytes.NewReader(b))
	} else {
		ret, err = http.NewRequestWithContext(ctx, method, url, bytes.NewReader(b))
	}

	if data != nil {
		ret.Header.Set("content-type", "application/json")
	}
	for h, v := range header {
		ret.Header.Set(h, v)
	}

	return
}

func doReq[RESP any](client *http.Client, req *http.Request) (ret RESP, err error) {
	resp, err := client.Do(req)
	if err != nil {
		return ret, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ret, NewErr(http.StatusInternalServerError, "error read body", nil)
	}
	if resp.StatusCode != 200 {
		return ret, NewErr(resp.StatusCode, string(b), nil)
	}

	if len(b) > 0 {
		err = json.Unmarshal(b, &ret)
	}
	return
}

func RequestCtx[RESP any](method, url string, query, header map[string]string, data any, ctx context.Context) (ret RESP, err error) {
	var client = http.DefaultClient
	req, err := buildReq(method, url, query, header, data, ctx)
	if err != nil {
		return
	}
	ret, err = doReq[RESP](client, req)
	return
}

func Request[RESP any](method, url string, query, header map[string]string, data any, timeout time.Duration) (ret RESP, err error) {
	var client = &http.Client{
		Timeout: timeout,
	}
	req, err := buildReq(method, url, query, header, data, nil)
	if err != nil {
		return
	}
	ret, err = doReq[RESP](client, req)
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

func GetQuery[T int | string | bool](ctx echo.Context, name string) (t T, err error) {
	switch any(t).(type) {
	case int:
		tt, e := strconv.Atoi(ctx.QueryParam(name))
		if e != nil {
			e = ErrBadRequest(e.Error())
			return t, e
		}
		t = any(tt).(T)
	case string:
		t = any(ctx.QueryParam(name)).(T)
	case bool:
		switch ctx.QueryParam(name) {
		case "true":
			t = any(true).(T)
		case "false":
			t = any(false).(T)
		default:
			err = errors.New("can only be true/false")
		}
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

func GetIDsByComma[T int | string](ctx echo.Context) (ret []T, err error) {
	ids := ctx.Param("ids")
	split := strings.Split(ids, ",")
	for _, s := range split {
		switch any(*new(T)).(type) {
		case int:
			n, err := strconv.Atoi(s)
			if err != nil {
				return ret, err
			}
			ret = append(ret, any(n).(T))
		case string:
			ret = append(ret, any(s).(T))
		}
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

type ReqKit struct {
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

func NewHttpTestKit(engine *echo.Echo, r RequestCfg) ReqKit {
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

	return ReqKit{
		Req:  req,
		Resp: resp,
		Ctx:  ctx,
	}
}

func MarshalResp[T any](resp *http.Response) (ret T, err error) {
	defer resp.Body.Close()
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return
	}
	err = Validator.Struct(&ret)
	return
}
