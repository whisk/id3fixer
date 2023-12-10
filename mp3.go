package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bogem/id3v2/v2"
	"github.com/rs/zerolog/log"
)

func fixMp3(src, dst string, fixFrames map[string]string, forced bool) error {
	log.Debug().Msgf("Fixing frames %v in file %s", fixFrames, src)
	if len(fixFrames) == 0 {
		return errors.New("nothing to fix")
	}
	// fail early
	if ok, _ := fileExists(src); !ok {
		return fmt.Errorf("%s does not exists", src)
	}

	tmpFile, err := os.CreateTemp("", "tmp*.mp3")
	if err != nil {
		return fmt.Errorf("failed creating temp file: %w", err)
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
			return fmt.Errorf("destination file %s already exists", dst)
		}
	}

	err = copyFileContents(src, tmpName)
	if err != nil {
		return fmt.Errorf("failed copying to temp file: %w", err)
	}

	tag, err := id3v2.Open(tmpName, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to read mp3 file: %w", err)
	}
	defer tag.Close()

	tag.SetDefaultEncoding(id3v2.EncodingUTF8)

	errorsCount := 0
	fixesCount := 0

	if _, ok := fixFrames["Comments"]; ok {
		actualComments := tag.GetFrames(tag.CommonID("Comments"))
		log.Debug().Msgf("Found %d comment tag(s)", len(actualComments))
		fixedComments := []id3v2.CommentFrame{}
		for i, comm := range actualComments {
			f := comm.(id3v2.CommentFrame)
			desc, err := fixEncoding(f.Description)
			if err != nil {
				log.Error().Msgf("Error converting comment frame %d description: %s", i, err)
				errorsCount += 1
				continue
			}
			text, err := fixEncoding(f.Text)
			if err != nil {
				log.Error().Msgf("Error converting comment frame %d text: %s", i, err)
				errorsCount += 1
				continue
			}
			newComm := id3v2.CommentFrame{
				Text:        text,
				Description: desc,
				Encoding:    id3v2.EncodingUTF8,
				Language:    f.Language,
			}
			log.Info().Msgf("Comment#%d %s -> %s", i, f.Text, text)
			fixesCount += 1
			fixedComments = append(fixedComments, newComm)
		}
		if len(fixedComments) > 0 {
			tag.DeleteFrames(tag.CommonID("Comments"))
			for _, c := range fixedComments {
				tag.AddCommentFrame(c)
			}
		}
	}
	for _, id := range fixFrames {
		if id[0] != 'T' {
			continue
		}
		actualFrames := tag.GetFrames(id)
		log.Debug().Msgf("Found %d %s tag(s)", len(actualFrames), id)
		fixedFrames := []id3v2.TextFrame{}
		for i, comm := range actualFrames {
			f := comm.(id3v2.TextFrame)
			text, err := fixEncoding(f.Text)
			if err != nil {
				log.Error().Msgf("Error converting text frame %s#%d: %s", id, i, err)
				errorsCount += 1
				continue
			}
			log.Info().Msgf("Frame %s#%d %s -> %s", id, i, f.Text, text)
			fixesCount += 1
			newText := id3v2.TextFrame{
				Text:     text,
				Encoding: id3v2.EncodingUTF8,
			}
			fixedFrames = append(fixedFrames, newText)
		}
		if len(fixedFrames) > 0 {
			tag.DeleteFrames(id)
			for _, t := range fixedFrames {
				tag.AddFrame(id, t)
			}
		}
	}

	if errorsCount > 0 {
		if !forced {
			return fmt.Errorf("got %d error(s) while fixing encoding and aborted", errorsCount)
		}
		log.Error().Msgf("Got %d errors(s) while fixing encoding, proceeding", errorsCount)
	}
	log.Debug().Msgf("Saving fixed file %s", dst)

	err = tag.Save()
	if err != nil {
		return fmt.Errorf("failed saving temp file: %w", err)
	}

	if dst != "" {
		err = copyFileSafe(tmpName, dst)
		if err != nil {
			return fmt.Errorf("failed creating output file: %w", err)
		}
	} else {
		// fix in-place
		backupFile := src + ".bak" + fmt.Sprint(time.Now().Unix())
		if ok, _ := fileExists(backupFile); ok {
			return errors.New("backup already exists")
		}
		err = copyFileContents(src, backupFile) // always make backups!
		if err != nil {
			return fmt.Errorf("failed creating a backup: %w", err)
		}
		err = copyFileContents(tmpName, src)
		if err != nil {
			return fmt.Errorf("failed to fix in-place: %w", err)
		}
	}
	log.Info().Msgf("Fixed %d frame(s)", fixesCount)

	return nil
}

func supportedMp3Frames() map[string]string {
	supportedFrames := make(map[string]string)
	for title, id := range id3v2.V23CommonIDs {
		if id == "COMM" || id[0] == 'T' {
			supportedFrames[title] = id
		}
	}
	return supportedFrames
}
