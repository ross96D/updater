package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGithubSignature(t *testing.T) {
	type Test struct {
		Token     []byte
		Hmac      []byte
		Signature []byte
	}
	testInputs := []Test{
		{
			Token:     []byte("sha256=757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"),
			Hmac:      []byte("sha256"),
			Signature: []byte("757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"),
		},
	}
	for _, test := range testInputs {
		hmac, signature, err := parseGithubSignature(test.Token)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(test.Hmac), string(hmac))
		assert.Equal(t, string(test.Signature), string(signature))
	}
}

func TestParseUserToken(t *testing.T) {
	type Test struct {
		Token  []byte
		Digest []byte
	}
	testInputs := []Test{
		{
			Token:  []byte("bearer 757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"),
			Digest: []byte("757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"),
		},
	}
	for _, test := range testInputs {
		digest, err := parseUserToken(test.Token)
		assert.Equal(t, nil, err)
		assert.Equal(t, string(test.Digest), string(digest))
	}
}
