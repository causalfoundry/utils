package util

import "net/http"

func ParseJwtToken(jwtToken string, jwtParsers ...JwtParser) (ret UserPayload, err error) {
	for _, parser := range jwtParsers {
		if ret, err = parser.TokenToPayload(jwtToken); err != nil {
			continue
		}
		return
	}
	err = NewErr(http.StatusUnauthorized, "all parser failed", nil)
	return
}
