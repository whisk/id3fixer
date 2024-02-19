package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixEncoding(t *testing.T) {
	actual, err := fixEncoding("ÐÀÎ Ãîâîðÿùàÿ êíèãà")
	assert.NoError(t, err)
	assert.Equal(t, actual, "РАО Говорящая книга")

	actual, err = fixEncoding(`"Âîêðóã ñâåòà"`)
	assert.NoError(t, err)
	assert.Equal(t, actual, `"Вокруг света"`, "should fix the mixture of ascii and cp1251")

	_, err = fixEncoding("А. и Б. Стругацкие")
	assert.ErrorContains(t, err, "rune not supported", "should fail on utf8")

	actual, err = fixEncoding("2005")
	assert.NoError(t, err)
	assert.Equal(t, actual, "2005", "should not change ascii")
}
