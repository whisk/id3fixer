package main

import (
	"fmt"
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/bogem/id3v2/v2"
	id3v1 "github.com/frolovo22/tag"
	"github.com/stretchr/testify/assert"
)

func checkV2GoldenFileIntegrity(t *testing.T, goldenFile string) {
	tag, err := id3v2.Open(goldenFile, id3v2.Options{Parse: true})
	if !assert.NoError(t, err) {
		t.Fatalf("failed to open %s, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}
	defer tag.Close()

	if !assert.Equal(t, "ÐÀÎ Ãîâîðÿùàÿ êíèãà", tag.GetTextFrame("TENC").Text) {
		t.Fatalf("test file %s seems incorrect, aborting. This is probably a bug in the tests or testdata", goldenFile)
	}
}

func checkV1GoldenFileIntegrity(t *testing.T, goldenFile string) {
}

func TestFixMp3(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkV2GoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err := fixMp3(goldenFile, tmpFileName, supportedV2Frames(), true)
	assert.NoError(t, err)

	tag, err := id3v2.Open(tmpFileName, id3v2.Options{Parse: true})
	assert.NoError(t, err)

	assert.Equal(t, "РАО Говорящая книга", tag.GetTextFrame("TENC").Text)
	assert.Equal(t, "Понедельник начинается в субботу", tag.Album())
	assert.Equal(t, "2005", tag.Year())
	assert.Equal(t, `"Вокруг света"`, tag.GetTextFrame("TCOP").Text)
}

func TestFixMp3Id3V1(t *testing.T) {
	goldenFile := "testdata/troika-id3v1.mp3"
	checkV1GoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("id3v1-%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err := fixMp3(goldenFile, tmpFileName, map[string]string{}, true)
	assert.NoError(t, err)

	tmpFh, err := os.Open(tmpFileName)
	assert.NoError(t, err, "failed to open tmp file. This is probably a bug in the tests or testdata")
	tag, err := id3v1.ReadID3v1(tmpFh)
	assert.NoError(t, err, "failed to read id3v1 tags. This is probably a bug in the tests or testdata")

	title, err := tag.GetTitle()
	assert.NoError(t, err)
	assert.Equal(t, "Gl. 1-1", title)
	artist, err := tag.GetArtist()
	assert.NoError(t, err)
	assert.Equal(t, "A. i  B. Strugatskie", artist)
	album, err := tag.GetAlbum()
	assert.NoError(t, err)
	assert.Equal(t, "Skazka o Troike", album)
	comment, err := tag.GetComment()
	assert.NoError(t, err)
	assert.Equal(t, "06:55, 44 100 Hz, Stereo, 19", comment)
}

func TestFixMp3_NoFixes(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkV2GoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	err := fixMp3(goldenFile, tmpFileName, map[string]string{}, true)
	assert.Error(t, err)
	assert.Equal(t, "failed fixing tags: no frames to fix given", err.Error())
}

func TestFixMp3_DestinationExists(t *testing.T) {
	goldenFile := "testdata/podenelnik-id3v2.mp3"
	checkV2GoldenFileIntegrity(t, goldenFile)

	tmpFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	defer os.Remove(tmpFileName)

	// Create a dummy destination file
	dstFileName := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
	file, err := os.Create(dstFileName)
	assert.NoError(t, err)
	file.Close()
	defer os.Remove(dstFileName)

	err = fixMp3(goldenFile, dstFileName, supportedV2Frames(), true)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("destination file %s already exists", dstFileName), err.Error())
}

func TestFixMp3_InPlaceFix(t *testing.T) {
    goldenFile := "testdata/podenelnik-id3v2.mp3"
    checkV2GoldenFileIntegrity(t, goldenFile)

    // Create a copy of the golden file
    tmpGoldenFile := path.Join(os.TempDir(), fmt.Sprintf("%d", rand.Uint64())+".mp3")
    err := copyFileContents(goldenFile, tmpGoldenFile)
    assert.NoError(t, err)
    defer os.Remove(tmpGoldenFile)

    backupFileName := tmpGoldenFile + ".bak"
    defer os.Remove(backupFileName)

    err = fixMp3(tmpGoldenFile, "", supportedV2Frames(), true)
    assert.NoError(t, err)

    tag, err := id3v2.Open(tmpGoldenFile, id3v2.Options{Parse: true})
    assert.NoError(t, err)
    defer tag.Close()

    assert.Equal(t, "РАО Говорящая книга", tag.GetTextFrame("TENC").Text)
    assert.Equal(t, "Понедельник начинается в субботу", tag.Album())
    assert.Equal(t, "2005", tag.Year())
    assert.Equal(t, `"Вокруг света"`, tag.GetTextFrame("TCOP").Text)
}
