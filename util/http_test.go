package util

import (
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
