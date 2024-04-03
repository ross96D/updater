//go:build windows

package taskservice

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByPath(t *testing.T) {
	ts, err := New()
	assert.Equal(t, nil, err)

	rt, err := ts.GetRegisteredTasks()
	assert.Equal(t, nil, err)

	for _, r := range rt {
		fmt.Printf("%s\n", r.Path)
	}
}
