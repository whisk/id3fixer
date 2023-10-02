package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// see https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file
// safely copy file with checks that destination does not exists
func copyFileSafe(src, dst string) (err error) {
	dstExists, _ := fileExists(dst)
	if dstExists {
		return fmt.Errorf("destination file %s already exists", dst)
	}

	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	// check source file
	srcStat, err := os.Stat(src)
	if err != nil {
		return
	}
	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("non-regular source file %s", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		closeDstErr := dstFile.Close()
		if err == nil {
			err = closeDstErr // if no error, try propagate close error if any
		}
	}()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return
	}

	err = dstFile.Sync()
	return
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err == nil {
		// file apparently exists
		return true, nil
	} else {
		// got error, let's see
		if errors.Is(err, os.ErrNotExist) {
			// file not exists, so no actual error here
			return false, nil
		} else {
			// other error
			return false, err
		}
	}
}
