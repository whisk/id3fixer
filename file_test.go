package main

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyFileSafe(t *testing.T) {
    srcFileName := path.Join(os.TempDir(), "srcFile.txt")
    dstFileName := path.Join(os.TempDir(), "dstFile.txt")
    defer os.Remove(srcFileName)
    defer os.Remove(dstFileName)

    // Create a source file
    err := os.WriteFile(srcFileName, []byte("Hello, World!"), 0644)
    assert.NoError(t, err)

    // Test copying the file safely
    err = copyFileSafe(srcFileName, dstFileName)
    assert.NoError(t, err)

    // Verify the contents of the destination file
    dstContents, err := os.ReadFile(dstFileName)
    assert.NoError(t, err)
    assert.Equal(t, "Hello, World!", string(dstContents))

    // Test copying to an existing destination file
    err = copyFileSafe(srcFileName, dstFileName)
    assert.Error(t, err)
    assert.Equal(t, "destination file "+dstFileName+" already exists", err.Error())
}
