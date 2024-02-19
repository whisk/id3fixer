package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/stretchr/testify/assert"
)

func TestFixMp3(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	tag, err := id3v2.Open(goldenFile, id3v2.Options{Parse: true})
	if !assert.NoError(t, err) {
		t.Fatalf("failed to open %s, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}
	defer tag.Close()

	if !assert.Equal(t, "ÐÀÎ Ãîâîðÿùàÿ êíèãà", tag.GetTextFrame("TENC").Text) {
		t.Fatalf("test file %s seems incorrect, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err = fixMp3(goldenFile, tmpFileName, supportedMp3Frames(), true)
	assert.NoError(t, err)

	tag, err = id3v2.Open(tmpFileName, id3v2.Options{Parse: true})
	assert.NoError(t, err)

	assert.Equal(t, "РАО Говорящая книга", tag.GetTextFrame("TENC").Text)
	assert.Equal(t, "Понедельник начинается в субботу", tag.Album())
	assert.Equal(t, "2005", tag.Year())
	assert.Equal(t, `"Вокруг света"`, tag.GetTextFrame("TCOP").Text)
}
