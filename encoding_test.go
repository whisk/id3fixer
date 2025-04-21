package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrokenCp1251ToUtf8(t *testing.T) {
	actual, err := brokenCp1251ToUtf8("ÐÀÎ Ãîâîðÿùàÿ êíèãà")
	assert.NoError(t, err)
	assert.Equal(t, actual, "РАО Говорящая книга")

	actual, err = brokenCp1251ToUtf8(`"Âîêðóã ñâåòà"`)
	assert.NoError(t, err)
	assert.Equal(t, actual, `"Вокруг света"`, "should fix the mixture of ascii and cp1251")

	_, err = brokenCp1251ToUtf8("А. и Б. Стругацкие")
	assert.ErrorContains(t, err, "rune not supported", "should fail on utf8")

	actual, err = brokenCp1251ToUtf8("2005")
	assert.NoError(t, err)
	assert.Equal(t, actual, "2005", "should not change ascii")
}

func TestCp1251ToTranslit(t *testing.T) {
	actual, err := cp1251ToTranslit(string([]byte{192, 46, 32, 232, 32, 32, 193, 46, 32, 209, 242, 240, 243, 227, 224, 246, 234, 232, 229}), 80)
	assert.NoError(t, err)
	assert.Equal(t, "A. i  B. Strugatskie", actual, "should transliterate")

	actual, err = cp1251ToTranslit("06:55, 44 100 Hz, Stereo, 19", 80)
	assert.NoError(t, err)
	assert.Equal(t, "06:55, 44 100 Hz, Stereo, 19", actual, "should not change ascii")
}

func TestTruncateUtf8(t *testing.T) {
	actual := truncateUtf8("А. и Б. Стругацкие", 80)
	assert.Equal(t, "А. и Б. Стругацкие", actual, "should not truncate")

	actual = truncateUtf8("А. и Б. Стругацкие", 29)
	assert.Equal(t, "А. и Б. Стругацки", actual, "should truncate")

	actual = truncateUtf8("А. и Б. Стругацкие", 30)
	assert.Equal(t, "А. и Б. Стругацки", actual, "should truncate with respect to utf8")
}
