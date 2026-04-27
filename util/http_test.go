package util

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetPayload(t *testing.T) {
	type greet struct {
		Target string `json:"name" validate:"required"`
		Greet  string `json:"greet" validate:"required"`
	}

	e := echo.New()
	t.Run("test single struct", func(t *testing.T) {

		kit := NewHttpTestKit(e, RequestCfg{
			Payload: greet{Target: "human", Greet: "hello"},
		})

		g, err := GetPayload[greet](kit.Ctx)
		assert.Nil(t, err)
		assert.Equal(t, g.Greet, "hello")
		assert.Equal(t, g.Target, "human")

		g2, err := GetPayload[greet](kit.Ctx)
		assert.Nil(t, err)
		assert.Equal(t, g2.Greet, "hello")
		assert.Equal(t, g2.Target, "human")
	})

	t.Run("test slice of struct", func(t *testing.T) {
		payload := []greet{
			{
				Target: "a",
				Greet:  "b",
			},
			{
				Target: "a",
				Greet:  "b",
			},
			{
				Target: "a",
				Greet:  "b",
			},
		}
		kit := NewHttpTestKit(e, RequestCfg{
			Payload: payload,
		})

		g, err := GetPayloads[greet](kit.Ctx)
		assert.Nil(t, err)
		assert.Equal(t, payload, g)

		// try again
		g, err = GetPayloads[greet](kit.Ctx)
		assert.Nil(t, err)
		assert.Equal(t, payload, g)
	})

	t.Run("test fail", func(t *testing.T) {
		kit := NewHttpTestKit(e, RequestCfg{
			Payload: greet{Target: "", Greet: "hello"},
		})

		g, err := GetPayload[greet](kit.Ctx)
		assert.NotNil(t, err)
		assert.Equal(t, err.(Err).Code, http.StatusBadRequest)
		assert.Equal(t, g.Greet, "hello")
		assert.Equal(t, g.Target, "")
	})

	t.Run("test slice fail includes index", func(t *testing.T) {
		payload := []greet{
			{Target: "a", Greet: "hello"},
			{Target: "", Greet: "hello"},
			{Target: "c", Greet: "hello"},
		}
		kit := NewHttpTestKit(e, RequestCfg{
			Payload: payload,
		})

		g, err := GetPayloads[greet](kit.Ctx)
		assert.NotNil(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
		assert.Contains(t, httpErr.Message, "payload[1]")
		assert.Equal(t, payload, g)
	})
}

func TestBind(t *testing.T) {
	e := echo.New()

	type req struct {
		Name string `query:"name" validate:"required"`
		Age  int    `query:"age" validate:"required"`
	}

	var req1, req2 req
	kit := NewHttpTestKit(e, RequestCfg{
		Method: "GET",
		QueryParams: map[string]string{
			"name": "tester",
			"age":  "3",
		},
	})

	err := Bind(kit.Ctx, &req1)
	assert.Nil(t, err)

	kit = NewHttpTestKit(e, RequestCfg{
		Method: "GET",
		QueryParams: map[string]string{
			"badname": "tester",
			"age":     "3",
		},
	})

	err = Bind(kit.Ctx, &req2)
	assert.NotNil(t, err)
}

func TestBuildReqEncodesQuery(t *testing.T) {
	req, err := buildReq(
		http.MethodGet,
		"http://example.com/search?existing=1",
		map[string]string{
			"q":   "a b&c",
			"tag": "x/y",
		},
		nil,
		nil,
		nil,
	)
	assert.Nil(t, err)
	assert.Equal(t, "existing=1&q=a+b%26c&tag=x%2Fy", req.URL.RawQuery)
	assert.Equal(t, "a b&c", req.URL.Query().Get("q"))
	assert.Equal(t, "x/y", req.URL.Query().Get("tag"))
}

func TestBuildReqReturnsMarshalError(t *testing.T) {
	type badPayload struct {
		Ch chan int `json:"ch"`
	}

	_, err := buildReq(
		http.MethodPost,
		"http://example.com/search",
		nil,
		nil,
		badPayload{Ch: make(chan int)},
		nil,
	)
	assert.NotNil(t, err)
}

type errReadCloser struct{}

func (errReadCloser) Read(_ []byte) (int, error) {
	return 0, errors.New("read failure")
}

func (errReadCloser) Close() error {
	return nil
}

func TestUnmarshalRespReturnsReadError(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(errReadCloser{}),
	}

	_, err := UnmarshalResp[map[string]any](resp)
	assert.ErrorContains(t, err, "read failure")
}
