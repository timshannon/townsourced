// Townsourced
// Copyright 2015 Tim Shannon. All rights reserved.

package app

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"strings"
	"unicode"
)

//Random returns a random, url safe value of the bit length passed in
func Random(bits int) string {
	result := make([]byte, bits/8)
	_, err := io.ReadFull(rand.Reader, result)
	if err != nil {
		panic(fmt.Sprintf("Error generating random values: %v", err))
	}

	return base64.RawURLEncoding.EncodeToString(result)
}

// for validating and making usernames, town names, or
// anything that needs to be url safe, case insensitive and user readable
type urlify string

// test if is valid urlForm
func (u urlify) is() bool {
	if u == "" {
		return false
	}
	for _, c := range u {
		if !u.validRune(c) {
			return false
		}
	}
	return true
}

func (u urlify) validRune(c rune) bool {
	return unicode.IsLetter(c) || unicode.IsNumber(c) || c == '-'
}

func (u urlify) make() string {
	for i, c := range u {
		if !u.validRune(c) {
			//TODO: benchmark vs buffer
			u = u[:i] + "-" + u[i+1:]
		}
	}

	return strings.ToLower(string(u))
}

func round(n float64) int {
	return int(n + math.Copysign(0.5, n))
}
