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
	}

	for _, tc := range testCases {
		res := NewAspectRatioFrom(tc.width, tc.height)
		assert.Equal(t, tc.expected, res)
	}
}
