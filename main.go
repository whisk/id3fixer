package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/bogem/id3v2/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
)

// TODO
// + file copy
// + temp file creation
// + logging/debug levels
// + cmdline options
// + push to github
// + modularity
// + error wrapping/unwrapping

var srcName = "sample.mp3"
var dstName = "sample-result.mp3"

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	tmpFile, err := os.CreateTemp("", "tmp*.mp3")
	if err != nil {
		log.Error().Msgf("Error creating temp file: %s", err)
		return
	}
	tmpName := tmpFile.Name()
	defer func() {
		tmpErr := tmpFile.Close()
		if tmpErr != nil {
			log.Error().Msgf("Error closing temp file: %s", tmpErr)
		}
		tmpErr = os.Remove(tmpName)
		if tmpErr != nil {
			log.Error().Msgf("Error removing temp file: %s", tmpErr)
		}
	}()

	// fail early
	dstExists, err := fileExists(dstName)
	if err != nil {
		log.Error().Msgf("error accessing destination file: %s", err)
		return
	}
	if dstExists {
		log.Error().Msgf("destination file %s already exists\n", dstName)
		return
	}

	err = copyFileContents(srcName, tmpName)
	if err != nil {
		log.Error().Msgf("Error copying to temp file: %s", err)
		return
	}

	tag, err := id3v2.Open(tmpName, id3v2.Options{Parse: true})
	if err != nil {
		log.Error().Msgf("Failed to read mp3 file: %s", err)
		return
	}
	defer tag.Close()

	title := tag.Title()
	artist := tag.Artist()
	album := tag.Album()
	log.Debug().Msgf("Raw title: '%s', raw artist: '%s'", title, artist)
	fixedTitle, err := fixEncoding(title)
	if err != nil {
		log.Error().Msgf("Error converting title: %s", err)
	}
	fixedArtist, err := fixEncoding(artist)
	if err != nil {
		log.Error().Msgf("Error converting title: %s", err)
	}
	fixedAlbum, err := fixEncoding(album)
	if err != nil {
		log.Error().Msgf("Error converting album: %s", err)
	}

	log.Info().Msgf("Fixed title: '%s', fixed artist: '%s'", fixedTitle, fixedArtist)
	tag.SetDefaultEncoding(id3v2.EncodingUTF8)
	tag.SetTitle(fixedTitle)
	tag.SetArtist(fixedArtist)
	tag.SetAlbum(fixedAlbum)

	err = tag.Save()
	if err != nil {
		log.Error().Msgf("Error saving temp file: %s", err)
		return
	}

	err = copyFile(tmpName, dstName)
	if err != nil {
		log.Error().Msgf("Error creating output file: %s", err)
		return
	}
}

// see https://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file
func copyFile(src, dst string) (err error) {
	// check source file
	srcStat, err := os.Stat(src)
	if err != nil {
		return
	}
	if !srcStat.Mode().IsRegular() {
		return fmt.Errorf("non-regular source file %s", src)
	}
	dstExists, _ := fileExists(dst)
	if dstExists {
		return fmt.Errorf("destination file %s already exists", dst)
	}

	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
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

func fixEncoding(s string) (string, error) {
	log.Debug().Msg("Bytes: \n" + formatBytes([]byte(s)))

	encoder := charmap.Windows1252.NewEncoder()
	res, err := encoder.String(s)
	if err != nil {
		return "", err
	}
	log.Debug().Msg("1 pass: \n" + formatBytes([]byte(res)))

	decoder := charmap.Windows1251.NewDecoder()
	res, err = decoder.String(res)
	if err != nil {
		return "", err
	}
	log.Debug().Msg("2 pass: \n" + formatBytes([]byte(res)))

	return res, nil
}

func formatBytes(arr []byte) string {
	s := ""
	for _, b := range arr {
		s += fmt.Sprintf("%08b ", b)
	}
	return s
}

func fileExists(name string) (bool, error) {
	_, err := os.Stat(dstName)
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
