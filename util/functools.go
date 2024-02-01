package util

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"time"
)

func Dedup[T any](arr []T) (ret []T) {
	var dedup = make(map[string]struct{})
	for _, t := range arr {
		key := fmt.Sprint(t)
		if _, ok := dedup[key]; !ok {
			ret = append(ret, t)
			dedup[key] = struct{}{}
		}
	}
	return
}

func Last[T any](arr []T) (ret T) {
	if len(arr) == 0 {
		return
	}

	return arr[len(arr)-1]
}

func NotNullCount(ts ...interface{}) int {
	var cnt int
	for _, t := range ts {
		if !reflect.ValueOf(t).IsNil() {
			cnt++
		}
	}
	return cnt
}

func MaxBy[T any, R int | float64](arr []T, maxBy func(t T) R) (max R) {
	for _, elem := range arr {
		if maxBy(elem) > max {
			max = maxBy(elem)
		}
	}
	return
}

func UniqueOf[T, R any](arr []T, fn func(T) R) (ret []R) {
	var tmp = make(map[string]bool)
	for _, a := range arr {
		r := fn(a)
		if tmp[fmt.Sprint(r)] {
			continue
		}
		ret = append(ret, r)
		tmp[fmt.Sprint(r)] = true
	}
	return
}

func UniqueBy[T any](arr []T, fn func(T) string) (ret []T) {
	var m = make(map[string]bool)
	for _, a := range arr {
		key := fn(a)
		if !m[key] {
			ret = append(ret, a)
			m[key] = true
		}
	}
	return
}

func Unique[T any](arr []T) (ret []T) {
	var dedup = make(map[string]bool)
	for _, t := range arr {
		var key string
		switch v := any(t).(type) {
		case time.Time:
			key = v.Format(time.RFC3339)
		default:
			key = fmt.Sprint(t)
			// simple handling for now
		}

		if dedup[key] {
			continue
		}
		dedup[key] = true
		ret = append(ret, t)
	}
	return
}

func Filterm[K comparable, V any](m map[K]V, fn func(k K, v V) bool) (ret map[K]V) {
	ret = make(map[K]V)
	for k, v := range m {
		if fn(k, v) {
			ret[k] = v
		}
	}
	return
}

func FlatMap[T, R any](arr []T, fn func(T) []R) (ret []R) {
	for _, a := range arr {
		ret = append(ret, fn(a)...)
	}
	return
}

func Mapm[K comparable, V, O any](m map[K]V, fn func(k K, v V) O) (ret []O) {
	ret = make([]O, len(m))
	var cnt int
	for k, v := range m {
		ret[cnt] = fn(k, v)
		cnt++
	}
	return
}

func Flatten[T any](arr [][]T) (ret []T) {
	for _, a := range arr {
		ret = append(ret, a...)
	}
	return
}

func ForEach[T any](arr []T, sideEffectFunc func(t T)) {
	for i := range arr {
		sideEffectFunc(arr[i])
	}
}

func MapInPlaceE[T any](arr []T, mapFunc func(t T) (T, error)) (err error) {
	for i := range arr {
		if arr[i], err = mapFunc(arr[i]); err != nil {
			return err
		}
	}
	return
}

func MapInPlace[T any](arr []T, mapFunc func(t T) T) {
	for i := range arr {
		arr[i] = mapFunc(arr[i])
	}
}

func MapE[T, R any](arr []T, mapFunc func(t T) (R, error)) (ret []R, err error) {
	ret = make([]R, len(arr))
	for i := range arr {
		if ret[i], err = mapFunc(arr[i]); err != nil {
			return
		}
	}
	return
}

func Map[T, R any](arr []T, mapFunc func(t T) R) (ret []R) {
	ret = make([]R, len(arr))
	for i := range arr {
		ret[i] = mapFunc(arr[i])
	}
	return
}

func Head[T any, K comparable](arr []T) (ret T) {
	if len(arr) == 0 {
		return
	}
	return arr[0]
}

func Tail[T any](arr []T) (ret T) {
	if len(arr) == 0 {
		return
	}
	return arr[len(arr)-1]
}

func Paginate[T any](arr []T, page Page) []T {
	if page.Offset() >= uint64(len(arr)) {
		return []T{}
	}

	offset := int(page.Offset())
	limit := int(page.Limit())
	last := Min(offset+limit, len(arr))

	return arr[offset:last]
}

func IndexOf[T any](arr []T, equalFunc func(t T) bool) int {
	for i, t := range arr {
		if equalFunc(t) {
			return i
		}
	}
	return -1
}

func Sum[T int | float32 | float64](arr []T) (ret T) {
	ret = 0
	for _, t := range arr {
		ret += t
	}
	return
}

func Avg[T int | float32 | float64](arr []T) (ret T) {
	return Sum(arr) / T(len(arr))
}

func FilterMap[T, R any](arr []T, fn func(t T) (R, bool)) (ret []R) {
	for _, t := range arr {
		if r, ok := fn(t); ok {
			ret = append(ret, r)
		}
	}
	return
}

func Filter[T any](arr []T, filterFunc func(t T) bool) (ret []T) {
	for _, t := range arr {
		if filterFunc(t) {
			ret = append(ret, t)
		}
	}
	return ret
}

func In[T any](in T, space []T) bool {
	for _, s := range space {
		if fmt.Sprint(s) == fmt.Sprint(in) {
			return true
		}
	}
	return false
}

func FirstN[T any](arr []T, n int) (ret []T) {
	return arr[:Min(n, len(arr))]
}

func LastN[T any](arr []T, n int) (ret []T) {
	return arr[Max(0, len(arr)-n):]
}

func PickRandom[T any](choices []T) (ret T) {
	if len(choices) == 0 {
		return
	}
	return choices[rand.Intn(len(choices))]
}

// can use relative tolerance, but for now absolute tolerance is sufficient
func SimpleFloatEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

// remove all items find in `except` from `base`, naive implementation with string equality
func Except[T any](base, except []T) (ret []T) {
	exceptMap := make(map[string]struct{})
	for i := range except {
		exceptMap[ToStr(except[i])] = struct{}{}
	}

	for i := range base {
		if _, ok := exceptMap[ToStr(base[i])]; !ok {
			ret = append(ret, base[i])
		}
	}
	return
}

func Shuffle[T any](arr []T) {
	rand.Shuffle(len(arr), func(i, j int) {
		arr[i], arr[j] = arr[j], arr[i]
	})
}

func FromAnys[T any](arr []any) []T {
	return Map(arr, func(t any) T { return t.(T) })
}

func ToAnys[T any](arr []T) []any {
	return Map(arr, func(t T) any { return any(t) })
}

func OneOf[T any](t T, arr []T) bool {
	for _, a := range arr {
		if fmt.Sprint(a) == fmt.Sprint(t) {
			return true
		}
	}
	return false
}

func ToMap[T comparable](arr []T) (ret map[T]bool) {
	ret = make(map[T]bool)
	for i := range arr {
		ret[arr[i]] = true
	}
	return
}

func SetIntersect[T comparable](a, b map[T]bool) (ret map[T]bool) {
	ret = make(map[T]bool)
	for i := range a {
		if b[i] {
			ret[i] = true
		}
	}
	return
}

func SetDifference[T comparable](a, b map[T]bool) (ret map[T]bool) {
	ret = make(map[T]bool)
	for i := range a {
		if !b[i] {
			ret[i] = true
		}
	}
	return
}

func SetValues[T comparable](a map[T]bool) (ret []T) {
	for i := range a {
		ret = append(ret, i)
	}
	return
}

func CollectBy[T any, R int | string](arr []T, fn func(T) R) (ret map[R]T) {
	ret = make(map[R]T)
	for _, a := range arr {
		ret[fn(a)] = a
	}
	return
}
func Min[T int | float32 | float64 | uint64](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func Max[T int | float32 | float64](a, b T) T {
	if a < b {
		return b
	}
	return a
}

func IndexesExcept[T any](sub, super []T) (ret []int) {
	for i := range super {
		var find bool
		for j := range sub {
			if fmt.Sprint(sub[j]) == fmt.Sprint(super[i]) {
				find = true
			}
		}
		if !find {
			ret = append(ret, i)
		}
	}
	return
}

func IndexesOf[T any](sub, super []T) (ret []int, valid bool) {
	for i := range sub {
		var find bool
		for j := range super {
			if fmt.Sprint(sub[i]) == fmt.Sprint(super[j]) {
				find = true
				ret = append(ret, j)
			}
		}

		if find {
			continue
		}
		return
	}

	valid = true
	return
}

// whether at least one of the elements in a is present in b
func ContainsAny[T any](a, b []T) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for i := range a {
		if Contains(a[i], b) {
			return true
		}
	}
	return false
}

// b contains all elements in a
func ContainsAll[T any](a, b []T) bool {
	if len(a) == 0 {
		return false
	}
	if len(b) == 0 {
		return true
	}

	for i := range a {
		if !Contains(a[i], b) {
			return false
		}
	}
	return true
}

func ContainsBy[T any](a T, b []T, fn func(a, b T) bool) bool {
	for i := range b {
		if fn(a, b[i]) {
			return true
		}
	}
	return false
}

// check if b has at least one of element in a
// equality is simply by converting the type to string
func Contains[T any](a T, b []T) bool {
	sa := fmt.Sprint(a)
	for i := range b {
		if sa == fmt.Sprint(b[i]) {
			return true
		}
	}
	return false
}

func SliceEqual[T any](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if fmt.Sprint(a[i]) != fmt.Sprint(b[i]) {
			return false
		}
	}
	return true
}

// HasIntersect
// check if another slices has at least one element in the base slice
// equality check is done by converting element to string
func HasIntersect[T any](base []T, another []T, emptyTrue bool) bool {
	if len(another) == 0 && len(base) != 0 {
		return false
	}

	if emptyTrue && len(base) == 0 {
		return true
	}

	for _, b := range base {
		for _, a := range another {
			if fmt.Sprint(b) == fmt.Sprint(a) {
				return true
			}
		}
	}
	return false
}

// this implementation didn't take consider if a, b has duplicates
func Intersect[T string | int | float64](a, b []T) (ret []T) {
	var ma = make(map[T]bool)
	for _, aa := range a {
		ma[aa] = true
	}

	for _, bb := range b {
		if ma[bb] {
			ret = append(ret, bb)
		}
	}
	return
}

