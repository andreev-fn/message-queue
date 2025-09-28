package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewQueueName_Valid(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{name: "simple alpha", in: "abc"},
		{name: "alphanumeric", in: "queue1"},
		{name: "two segments", in: "abc.def"},
		{name: "three segments", in: "abc.def.ghi"},
		{name: "mixed segments", in: "a.3"},
		{name: "all numbers", in: "1.2.3"},
		{name: "underscore allowed", in: "a_b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewQueueName(tc.in)
			require.NoError(t, err)
			require.Equal(t, tc.in, got.String())
		})
	}
}

func TestNewQueueName_Invalid(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{name: "empty string", in: ""},
		{name: "empty part between dots", in: "a..b"},
		{name: "leading dot empty part", in: ".a"},
		{name: "trailing dot empty part", in: "a."},
		{name: "uppercase not allowed", in: "Abc"},
		{name: "dash not allowed", in: "a-b"},
		{name: "space not allowed", in: "a b"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := NewQueueName(tc.in)
			require.Error(t, err)
		})
	}
}
