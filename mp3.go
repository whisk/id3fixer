package main

import (
	"fmt"
	"os"

	"github.com/bogem/id3v2/v2"
	"github.com/rs/zerolog/log"
)

func fixMp3(src, dst string, fixFrames map[string]string, forced bool) error {
	log.Debug().Msgf("Fixing frames %v in file %s", fixFrames, src)
	if len(fixFrames) == 0 {
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

	errorsCount := 0

	if _, ok := fixFrames["Comments"]; ok {
		actualComments := tag.GetFrames(tag.CommonID("Comments"))
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
			log.Debug().Msgf("Comment#%d %s -> %s", i, f.Text, text)
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
		fixedFrames := []id3v2.TextFrame{}
		for i, comm := range actualFrames {
			f := comm.(id3v2.TextFrame)
			text, err := fixEncoding(f.Text)
			if err != nil {
				log.Error().Msgf("Error converting text frame %s#%d: %s", id, i, err)
				errorsCount += 1
				continue
			}
			log.Debug().Msgf("Frame %s#%d %s -> %s", id, i, f.Text, text)
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
			return fmt.Errorf("Got %d error(s) while fixing encoding, aborting", errorsCount)
		}
		log.Error().Msgf("Got %d errors(s) while fixing encoding, proceeding", errorsCount)
	}
	log.Debug().Msgf("Saving fixed file %s", dst)

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

func supportedMp3Frames() map[string]string {
	supportedFrames := make(map[string]string)
	for title, id := range id3v2.V23CommonIDs {
		if id == "COMM" || id[0] == 'T' {
			supportedFrames[title] = id
		}
	}
	return supportedFrames
}
