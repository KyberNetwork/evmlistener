package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	expect := "test"

	data, err := Encode(expect)
	require.NoError(t, err)

	var v string
	err = Decode(data, &v)
	require.NoError(t, err)

	assert.Equal(t, expect, v)
}

func TestFormatKey(t *testing.T) {
	tests := []struct {
		args   []string
		expect string
	}{
		{
			args:   []string{"prefix-key:", "123456"},
			expect: "prefix-key:123456",
		},
		{
			args:   []string{"prefix_key:", "654321"},
			expect: "prefix_key-654321",
		},
	}

	for _, test := range tests {
		key := FormatKey(test.args...)
		assert.Equal(t, test.expect, key)
	}
}
