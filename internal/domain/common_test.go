package domain

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAspectRatioFrom(t *testing.T) {
	testCases := []struct {
		width    int
		height   int
		expected AR
	}{
		{width: 275, height: 183, expected: AR{275, 183}},
		{width: 100, height: 100, expected: AR{1, 1}},
		{width: 100, height: 200, expected: AR{1, 2}},
		{width: 0, height: 100, expected: AR{Width: 0, Height: 0}},
		{width: 100, height: 0, expected: AR{Width: 0, Height: 0}},
		{width: 0, height: 0, expected: AR{0, 0}},
	}

	for _, tc := range testCases {
		res := NewAspectRatioFrom(tc.width, tc.height)
		assert.Equal(t, tc.expected, res)
	}
}

func TestGreatCommonFactor(t *testing.T) {
	testCases := []struct {
		a        int
		b        int
		expected int
	}{
		{a: 10, b: 0, expected: 10},
		{a: 0, b: 0, expected: 0},
		{a: 275, b: 183, expected: 1},
		{a: 6, b: 3, expected: 3},
	}
	for _, tc := range testCases {
		res := GreatCommonFactor(tc.a, tc.b)
		assert.Equal(t, tc.expected, res)
	}
}
