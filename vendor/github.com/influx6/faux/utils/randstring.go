package utils

import (
	"crypto/rand"
	"math"
	mrand "math/rand"
	"strings"
)

// RandomInt returns a random number between the provided min-max range.
func RandomInt(min, max int64) int64 {
	return int64(math.Floor(float64(mrand.Int63()*((max-min)+1)))) + min
}

// RandomFloat returns a random number between the provided min-max range.
func RandomFloat(min, max float64) float64 {
	return math.Floor(mrand.Float64()*((max-min)+1)) + min
}

// RandString generates a set of random numbers of a set length
func RandString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// RandText generates random string based on type (Alphanum, Alpha, Number).
func RandText(strSize int, randType string) string {
	var dictionary string

	randType = strings.ToLower(randType)

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytes = make([]byte, strSize)
	rand.Read(bytes)

	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}

	return string(bytes)
}
