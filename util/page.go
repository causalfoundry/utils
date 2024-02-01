package util

import (
	"fmt"
	"math"
	"strconv"

	"github.com/labstack/echo/v4"
)

var NoPagination = Page{Off: true}

type PaginatedResult struct {
	Total int `json:"total"`
	Data  any `json:"data"`
}

// Error implements error
func (PaginatedResult) Error() string {
	panic("unimplemented")
}

type Page struct {
	Page    int `query:"page"`
	PerPage int `query:"per_page"`
	Off     bool
}

func (p Page) Offset() uint64 {
	if p.Off {
		return 0
	}
	return uint64(p.Page * p.PerPage)
}

func (p Page) Limit() uint64 {
	if p.Off {
		return math.MaxInt64
	}
	return uint64(p.PerPage)
}

// only valid for database array, has off by one error for go array
func (p Page) DBSlice(col string) string {
	if p.Off {
		return fmt.Sprintf("%s[0:]", col)
	}
	return fmt.Sprintf("%s[%d:%d]", col, p.Page*p.PerPage, (p.Page+1)*p.PerPage)
}

func (p Page) GoSliceStartEnd(slen int) (int, int) {
	if p.Off {
		return 0, slen
	}

	start := p.Page * p.PerPage
	if start > slen {
		start = slen
	}

	end := start + p.PerPage
	if end > slen {
		end = slen
	}

	return start, end
}

func GetPagination(ctx echo.Context) (ret Page, err error) {
	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil {
		err = ErrBadRequest(err.Error())
		return
	}
	perPage, err := strconv.Atoi(ctx.QueryParam("per_page"))
	if err != nil {
		err = ErrBadRequest(err.Error())
		return
	}

	if page == -1 && perPage == -1 {
		ret = NoPagination
		return
	}
	ret.Page = page
	ret.PerPage = perPage
	return
}
