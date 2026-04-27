package util

import (
	"errors"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

type fakeJwtParser struct {
	payload UserPayload
	err     error
}

func (f fakeJwtParser) TokenToPayload(_ string) (UserPayload, error) {
	return f.payload, f.err
}

func TestMS(t *testing.T) {
}

func TestNewMicrosoftJwtParser(t *testing.T) {
	t.Run("require audience or issuer", func(t *testing.T) {
		_, err := NewMicrosoftJwtParser(MicrosoftJwtParserConfig{})
		assert.ErrorContains(t, err, "allowed audience or issuer")
	})

	t.Run("accept audience config", func(t *testing.T) {
		_, err := NewMicrosoftJwtParser(MicrosoftJwtParserConfig{
			AllowedAudiences: []string{"client-id"},
		})
		assert.NoError(t, err)
	})

	t.Run("accept issuer config", func(t *testing.T) {
		_, err := NewMicrosoftJwtParser(MicrosoftJwtParserConfig{
			AllowedIssuers: []string{"https://issuer.example.com"},
		})
		assert.NoError(t, err)
	})
}

func TestMicrosoftJwtParserValidateClaims(t *testing.T) {
	parser, err := NewMicrosoftJwtParser(MicrosoftJwtParserConfig{
		AllowedAudiences: []string{"client-id"},
		AllowedIssuers:   []string{"https://issuer.example.com"},
	})
	assert.NoError(t, err)

	t.Run("accept matching audience and issuer", func(t *testing.T) {
		err := parser.validateClaims(jwt.MapClaims{
			"aud": "client-id",
			"iss": "https://issuer.example.com",
		})
		assert.NoError(t, err)
	})

	t.Run("reject mismatched audience", func(t *testing.T) {
		err := parser.validateClaims(jwt.MapClaims{
			"aud": "other-client",
			"iss": "https://issuer.example.com",
		})
		assert.ErrorContains(t, err, "unexpected audience")
	})

	t.Run("reject mismatched issuer", func(t *testing.T) {
		err := parser.validateClaims(jwt.MapClaims{
			"aud": "client-id",
			"iss": "https://other-issuer.example.com",
		})
		assert.ErrorContains(t, err, "unexpected issuer")
	})

	t.Run("legacy constructor skips claim checks", func(t *testing.T) {
		parser := NewMicrosfotJwtParser(nil)
		err := parser.validateClaims(jwt.MapClaims{
			"aud": "other-client",
			"iss": "https://other-issuer.example.com",
		})
		assert.NoError(t, err)
	})
}

func TestParseJwtToken(t *testing.T) {
	t.Run("return first successful parser payload", func(t *testing.T) {
		payload, err := ParseJwtToken(
			"token",
			fakeJwtParser{err: errors.New("wrong issuer")},
			fakeJwtParser{payload: UserPayload{Username: "alice"}},
			fakeJwtParser{err: errors.New("should not be reached")},
		)
		assert.NoError(t, err)
		assert.Equal(t, "alice", payload.Username)
	})

	t.Run("return aggregated parser errors", func(t *testing.T) {
		payload, err := ParseJwtToken(
			"token",
			fakeJwtParser{err: errors.New("wrong issuer")},
			fakeJwtParser{err: errors.New("unexpected audience")},
		)
		assert.Equal(t, UserPayload{}, payload)
		assert.Error(t, err)

		appErr, ok := err.(Err)
		assert.True(t, ok)
		assert.Equal(t, 401, appErr.Code)
		assert.Contains(t, appErr.Msg, "all parser failed")
		assert.Contains(t, appErr.Msg, "wrong issuer")
		assert.Contains(t, appErr.Msg, "unexpected audience")
	})
}
