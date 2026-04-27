package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginatedResultError(t *testing.T) {
	var err error = PaginatedResult{Total: 3}
	assert.Equal(t, "paginated result: total=3", err.Error())
}
