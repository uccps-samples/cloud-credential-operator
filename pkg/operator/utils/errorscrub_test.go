package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorScrubber(t *testing.T) {
	cases := []struct {
		name string

		input    string
		expected string
	}{
		{
			name:     "aws request id",
			input:    "failed to grant creds: error syncing creds in mint-mode: AWS Error: LimitExceeded - LimitExceeded: Cannot exceed quota for UsersPerAccount: 5000\n\tstatus code: 409, request id: 0604c1a4-0a68-4d1a-b8e6-cdcf68176d71",
			expected: "failed to grant creds: error syncing creds in mint-mode: AWS Error: LimitExceeded - LimitExceeded: Cannot exceed quota for UsersPerAccount: 5000, status code: 409",
		},
		{
			name:     "request id mid",
			input:    "AWS Error: LimitExceeded - LimitExceeded: Cannot exceed quota for UsersPerAccount: 5000, request id: 0604c1a4-0a68-4d1a-b8e6-cdcf68176d71, something else",
			expected: "AWS Error: LimitExceeded - LimitExceeded: Cannot exceed quota for UsersPerAccount: 5000, something else",
		},
		{
			name:     "request id start",
			input:    "request id: 0604c1a4-0a68-4d1a-b8e6-cdcf68176d71, something else", // shouldn't really happen
			expected: ", something else",                                                 // not pretty but what I want to verify happens
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ErrorScrub(fmt.Errorf(test.input)))
		})
	}
}
