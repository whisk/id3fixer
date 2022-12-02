package main

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
	"github.com/rs/zerolog/log"
)

func fixMp3(src, dst string, fixTitle, fixArtist, fixAlbum bool) error {
	if !fixTitle && !fixArtist && !fixAlbum {
		return fmt.Errorf("Nothing to fix!")
	}

	tmpFile, err := os.CreateTemp("", "tmp*.mp3")
	if err != nil {
		return fmt.Errorf("Error creating temp file: %w", err)
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
	dstExists, err := fileExists(dst)
	if err != nil {
		return fmt.Errorf("error accessing destination file: %w", err)
	}
	if dstExists {
		return fmt.Errorf("destination file %s already exists\n", dst)
	}

	err = copyFileContents(src, tmpName)
	if err != nil {
		return fmt.Errorf("Error copying to temp file: %w", err)
	}

	tag, err := id3v2.Open(tmpName, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("Failed to read mp3 file: %w", err)
	}
	defer tag.Close()

	tag.SetDefaultEncoding(id3v2.EncodingUTF8)

	title := tag.Title()
	artist := tag.Artist()
	album := tag.Album()

	log.Debug().Msgf("Raw title: '%s', raw artist: '%s'", title, artist)
	if fixTitle {
		title, err = fixEncoding(title)
		if err != nil {
			log.Error().Msgf("Error converting title: %s", err)
		}
		tag.SetTitle(title)
	}

	if fixArtist {
		artist, err = fixEncoding(artist)
		if err != nil {
			log.Error().Msgf("Error converting artist: %s", err)
		}
		tag.SetArtist(artist)
	}

	if fixAlbum {
		album, err = fixEncoding(album)
		if err != nil {
			log.Error().Msgf("Error converting album: %s", err)
		}
		tag.SetAlbum(album)
	}

	log.Info().Msgf("Fixed title: '%s', fixed artist: '%s', fixed album: '%s'", title, artist, album)

	err = tag.Save()
	if err != nil {
		return fmt.Errorf("Error saving temp file: %w", err)
	}

	err = copyFile(tmpName, dst)
	if err != nil {
		return fmt.Errorf("Error creating output file: %w", err)
	}

	return nil
}
