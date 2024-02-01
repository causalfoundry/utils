package util

import (
	"fmt"
	"sort"
	"time"
)

func SortTime(ts []time.Time) {
	sort.Slice(ts, func(i, j int) bool {
		return ts[i].Before(ts[j])
	})
}

func SortAsStr[T any](arr []T) []T {
	sort.Slice(arr, func(i, j int) bool {
		return fmt.Sprint(arr[i]) < fmt.Sprint(arr[j])
	})
	return arr
}
