package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeTempFile(t *testing.T, content string) string {
	path := filepath.Join(t.TempDir(), "prefs.toml")
	err := os.WriteFile(path, []byte(content), 0600)
	require.Nil(t, err)

	return path
}
