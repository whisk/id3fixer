package main

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
	"github.com/rs/zerolog/log"
)

func fixMp3(src, dst string, fixTitle, fixArtist, fixAlbum, fixComments bool) error {
	if !fixTitle && !fixArtist && !fixAlbum && !fixComments {
		return fmt.Errorf("Nothing to fix!")
	}
	// fail early
	if ok, _ := fileExists(src); !ok {
		return fmt.Errorf("%s does not exists", src)
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

	if dst != "" {
		// fail early
		dstExists, err := fileExists(dst)
		if err != nil {
			return fmt.Errorf("error accessing destination file: %w", err)
		}
		if dstExists {
			return fmt.Errorf("destination file %s already exists\n", dst)
		}
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
	errorsCount := 0
	if fixTitle {
		title, err = fixEncoding(title)
		if err != nil {
			log.Error().Msgf("Error converting title: %s", err)
			errorsCount += 1
		}
		tag.SetTitle(title)
	}

	if fixArtist {
		artist, err = fixEncoding(artist)
		if err != nil {
			log.Error().Msgf("Error converting artist: %s", err)
			errorsCount += 1
		}
		tag.SetArtist(artist)
	}

	if fixAlbum {
		album, err = fixEncoding(album)
		if err != nil {
			log.Error().Msgf("Error converting album: %s", err)
			errorsCount += 1
		}
		tag.SetAlbum(album)
	}

	if fixComments {
		framesMap := tag.AllFrames()
		comments, ok := framesMap[tag.CommonID("Comments")]
		fixedComments := []id3v2.CommentFrame{}
		if ok {
			for i, comm := range comments {
				f := comm.(id3v2.CommentFrame)
				desc, err1 := fixEncoding(f.Description)
				text, err2 := fixEncoding(f.Text)
				if err1 != nil || err2 != nil {
					errorsCount += 1
					log.Error().Msgf("Error converting comment %d: %s, %s", i, err1, err2)
				} else {
					newComm := id3v2.CommentFrame{
						Text:        text,
						Description: desc,
						Encoding:    id3v2.EncodingUTF8,
						Language:    f.Language,
					}
					fixedComments = append(fixedComments, newComm)
				}
			}
			if len(fixedComments) > 0 {
				tag.DeleteFrames(tag.CommonID("Comments"))
				for _, c := range fixedComments {
					tag.AddCommentFrame(c)
				}
			}
		}
	}

	log.Info().Msgf("Fixed title: '%s', fixed artist: '%s', fixed album: '%s'", title, artist, album)
	if errorsCount > 0 {
		return fmt.Errorf("Got errors while fixing encoding, aborting")
	}

	err = tag.Save()
	if err != nil {
		return fmt.Errorf("Error saving temp file: %w", err)
	}

	if dst != "-" {
		err = copyFileSafe(tmpName, dst)
		if err != nil {
			return fmt.Errorf("Error creating output file: %w", err)
		}
	} else {
		// fix in-place
		err = copyFileContents(src, src+".bak") // always make backups!
		if err != nil {
			return fmt.Errorf("Error creating a backup: %w", err)
		}
		err = copyFileContents(tmpName, src)
		if err != nil {
			return fmt.Errorf("Error fixing in-place: %w", err)
		}
	}

	return nil
}
