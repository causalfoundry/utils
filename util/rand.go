package util

import (
	"math/rand"
	"strings"
)

func _alphabet() (ret string) {
	for i := 'a'; i <= 'z'; i++ {
		ret += string(i)
	}
	ret += strings.ToUpper(ret)
	return
}

var numeric = "0123456789"
var alphabets = _alphabet()
var alphanumeric = alphabets + numeric

func RandomIntBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func RandomString(low int, high int) (ret string) {
	return RandomAlphabets(RandomIntBetween(low, high), true)
}

func RandomInt(low int, high int) int {
	return rand.Intn(high-low) + low
}

func RandomFloat64(low float64, high float64) float64 {
	return rand.Float64()*(high-low) + low
}

func RandomBool() bool {
	return rand.Intn(2) == 1
}

func RandomAlphabets(l int, lowerCase bool) string {
	b := strings.Builder{}
	al := len(alphabets)
	r := NewRand()
	for i := 0; i < l; i++ {
		_ = b.WriteByte(alphabets[r.Intn(al)])
	}
	if lowerCase {
		return strings.ToLower(b.String())
	}
	return b.String()
}
