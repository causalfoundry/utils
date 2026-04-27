package util

import (
	"fmt"
	"net/http"
	"strings"
)

func ParseJwtToken(jwtToken string, jwtParsers ...JwtParser) (ret UserPayload, err error) {
	var parseErrs []string
	for _, parser := range jwtParsers {
		if ret, err = parser.TokenToPayload(jwtToken); err != nil {
			parseErrs = append(parseErrs, fmt.Sprintf("%T: %v", parser, err))
			continue
		}
		return
	}
	if len(parseErrs) != 0 {
		err = NewErr(http.StatusUnauthorized, fmt.Sprintf("all parser failed: %s", strings.Join(parseErrs, "; ")), nil)
		return
	}
	err = NewErr(http.StatusUnauthorized, "all parser failed", nil)
	return
}
