package models_test

import (
	"testing"

	"github.com/ross96D/updater/cmd/client/models"
	"github.com/stretchr/testify/require"
)

func TestUrlValidator(t *testing.T) {
	t.Run("fail", func(t *testing.T) {
		text := "192.0.1.1"
		_, err := models.URLValidator{}.ParseValidationItem(text)
		require.Error(t, err)

		text = "234.23.23.23"
		_, err = models.URLValidator{}.ParseValidationItem(text)
		require.Error(t, err)
	})
}
