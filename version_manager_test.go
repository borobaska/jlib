package jlib

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testJavaOptions = &JavaInstallOptions{
	Distribution:    []string{"zulu"},
	JDKVersion:      8,
	OperatingSystem: GetOS(),
	Architecture:    GetArch(),
}

func TestVersionManager(t *testing.T) {
	t.Run("Install", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}
		tmp := t.TempDir()
		vm := NewVersionManager(tmp)
		j, err := vm.Install(testJavaOptions)
		assert.NoError(t, err)
		assert.NotNil(t, j)
		assert.FileExists(t, path.Join(j.JavaDir, "meta.json"))
	})

	t.Run("List", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping test in short mode.")
		}
		tmp := t.TempDir()
		vm := NewVersionManager(tmp)
		j, err := vm.Install(&JavaInstallOptions{
			Distribution:    []string{"zulu"},
			JDKVersion:      8,
			OperatingSystem: []string{"linux"},
			Architecture:    []string{"amd64"},
		})
		assert.NoError(t, err)
		assert.NotNil(t, j)

		result, err := vm.List()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result)
	})
}
