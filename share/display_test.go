package share

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayTask(t *testing.T) {
	err := DisplayTasks()

	assert.Equal(t, err, nil)

}
