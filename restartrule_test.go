package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRestartRuleMarshalText(t *testing.T) {
	var tests = []struct {
		expected []byte
		value    RestartRule
	}{
		{[]byte("no"), No},
		{[]byte("on-error"), OnError},
	}

	for _, test := range tests {
		got, err := test.value.MarshalText()
		assert.NoError(t, err)
		assert.Equal(t, test.expected, got)
	}
}

func TestRestartRuleUnmarshalText(t *testing.T) {
	var tests = []struct {
		expected RestartRule
		value    []byte
	}{
		{No, []byte("no")},
		{OnError, []byte("on-error")},
	}

	for _, test := range tests {
		var r RestartRule
		err := r.UnmarshalText(test.value)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, r)
	}
}
