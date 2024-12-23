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

func checkGoldenFileIntegrity(t *testing.T, goldenFile string) {
	tag, err := id3v2.Open(goldenFile, id3v2.Options{Parse: true})
	if !assert.NoError(t, err) {
		t.Fatalf("failed to open %s, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}
	defer tag.Close()

	if !assert.Equal(t, "ÐÀÎ Ãîâîðÿùàÿ êíèãà", tag.GetTextFrame("TENC").Text) {
		t.Fatalf("test file %s seems incorrect, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}
}

func TestFixMp3(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkGoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err := fixMp3(goldenFile, tmpFileName, supportedMp3Frames(), true)
	assert.NoError(t, err)

	tag, err := id3v2.Open(tmpFileName, id3v2.Options{Parse: true})
	assert.NoError(t, err)

	assert.Equal(t, "РАО Говорящая книга", tag.GetTextFrame("TENC").Text)
	assert.Equal(t, "Понедельник начинается в субботу", tag.Album())
	assert.Equal(t, "2005", tag.Year())
	assert.Equal(t, `"Вокруг света"`, tag.GetTextFrame("TCOP").Text)
}

func TestFixMp3_NoFixes(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkGoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err := fixMp3(goldenFile, tmpFileName, map[string]string{}, true)
	assert.Error(t, err)
	assert.Equal(t, "nothing to fix", err.Error())
}

func TestFixMp3_DestinationExists(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkGoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	// Create a dummy destination file
	dstFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	file, err := os.Create(dstFileName)
	assert.NoError(t, err)
	file.Close()
	defer os.Remove(dstFileName)

	err = fixMp3(goldenFile, dstFileName, supportedMp3Frames(), true)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("destination file %s already exists", dstFileName), err.Error())
}

func TestFixMp3_InPlaceFix(t *testing.T) {
    goldenFile := "testdata/podenelnik-id3v2.mp3"
    checkGoldenFileIntegrity(t, goldenFile)

    // Create a copy of the golden file
    tmpGoldenFile := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
    err := copyFileContents(goldenFile, tmpGoldenFile)
    assert.NoError(t, err)
    defer os.Remove(tmpGoldenFile)

    backupFileName := tmpGoldenFile + ".bak"
    defer os.Remove(backupFileName)

    err = fixMp3(tmpGoldenFile, "", supportedMp3Frames(), true)
    assert.NoError(t, err)

    tag, err := id3v2.Open(tmpGoldenFile, id3v2.Options{Parse: true})
    assert.NoError(t, err)
    defer tag.Close()

    assert.Equal(t, "РАО Говорящая книга", tag.GetTextFrame("TENC").Text)
    assert.Equal(t, "Понедельник начинается в субботу", tag.Album())
    assert.Equal(t, "2005", tag.Year())
    assert.Equal(t, `"Вокруг света"`, tag.GetTextFrame("TCOP").Text)
}
