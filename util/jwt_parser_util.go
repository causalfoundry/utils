package util

import "net/http"

func ParseToken(token string, parsers ...JwtParser) (ret UserPayload, err error) {
	for _, parser := range parsers {
		if ret, err = parser.TokenToPayload(token); err != nil {
			continue
		}
		return
	}
	err = NewErr(http.StatusUnauthorized, "all parser failed", nil)
	return
}
