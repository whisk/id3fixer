package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/bogem/id3v2/v2"
	id3tag "github.com/frolovo22/tag"
	"github.com/rs/zerolog/log"
)

type Change struct {
	Old string
	New string
}

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

	err = fixTags(tmpName, fixFrames, forced)
	if err != nil {
		return fmt.Errorf("failed fixing tags: %w", err)
	}
	log.Debug().Msgf("Saving fixed file %s", dst)

	if dst != "" {
		err = copyFileSafe(tmpName, dst)
		if err != nil {
			return fmt.Errorf("failed creating output file: %w", err)
		}
	} else {
		// fix in-place
		backupFile := src + "." + fmt.Sprint(time.Now().Unix()) + ".bak"
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

	return nil
}

func fixTags(fileName string, fixFrames map[string]string, forced bool) error {
	file, err := id3tag.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read id3 tags to detect the version: %w", err)
	}
	if file.GetVersion() == id3tag.VersionID3v1 {
		return fixTagsV1(fileName, forced)
	} else if file.GetVersion() == id3tag.VersionID3v23 || file.GetVersion() == id3tag.VersionID3v24 {
		return fixTagsV23(fileName, fixFrames, forced)
	}

	return fmt.Errorf("unsupported id3 version" + file.GetVersion().String())
}

func fixTagsV1(fileName string, forced bool) error {
	fh, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("failed opening mp3 file for reading: %w", err)
	}
	defer fh.Close()
	file, err := id3tag.ReadID3v1(fh)
	if err != nil {
		return fmt.Errorf("failed to read id3v1 tags: %w", err)
	}
	type id3v1TagAccessor struct {
		Getter func() (string, error)
		Setter func(string) error
	}
	fn := make(map[string]id3v1TagAccessor)
	fn["Title"] = id3v1TagAccessor{Getter: file.GetTitle, Setter: file.SetTitle}
	fn["Artist"] = id3v1TagAccessor{Getter: file.GetArtist, Setter: file.SetArtist}
	fn["Album"] = id3v1TagAccessor{Getter: file.GetAlbum, Setter: file.SetAlbum}
	fn["Comment"] = id3v1TagAccessor{Getter: file.GetComment, Setter: file.SetComment}
	totalErrorsCount := 0
	totalFixedCount := 0
	for field, f := range fn {
		log.Debug().Msgf("found tag %s", field)
		val, err := f.Getter()
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to read tag %s", field)
			totalErrorsCount += 1
			continue
		}
		if val == "" {
			continue
		}
		fixedVal, err := fixEncodingCp1251(val, 30)
		if err != nil {
			log.Warn().Err(err).Msgf("Failed to fix tag %s", field)
			totalErrorsCount += 1
			continue
		}
		if fixedVal != val {
			log.Info().Msgf("Fixed tag %s: %s -> %s", field, val, fixedVal)
			totalFixedCount += 1
			err = f.Setter(fixedVal)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to set tag %s", field)
				totalErrorsCount += 1
			}
		}
	}
	if totalErrorsCount > 0 {
		if !forced {
			return fmt.Errorf("got %d error(s) while fixing encoding and aborted", totalErrorsCount)
		}
		log.Error().Msgf("Got %d errors(s) while fixing encoding, proceeding", totalErrorsCount)
	}

	fh.Close()
	fh, err = os.OpenFile(fileName, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("failed opening mp3 file for writing: %w", err)
	}
	defer fh.Close()

	err = file.Save(fh)
	if err != nil {
		return fmt.Errorf("failed to save id3v1 tags: %w", err)
	}
	log.Info().Msgf("Fixed %d tag(s)", totalFixedCount)

	return nil
}

func fixTagsV23(fileName string, fixFrames map[string]string, forced bool) error {
	tag, err := id3v2.Open(fileName, id3v2.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("failed to read mp3 file: %w", err)
	}
	defer tag.Close()
	tag.SetVersion(4)

	totalErrorsCount := 0
	totalFixedCount := 0
	for _, id := range fixFrames {
		actualFrames := tag.GetFrames(id)
		log.Debug().Msgf("Found %d %s tag(s)", len(actualFrames), id)
		fixedFrames := []id3v2.Framer{}
		fixesCount := 0
		for i, frame := range actualFrames {
			fixedFrame, fixes, err := fixFrame(frame)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to fix frame %s#%d, leaving it as is", id, i)
				totalErrorsCount += 1
				fixedFrames = append(fixedFrames, frame)
				continue
			}
			if fixes == nil {
				log.Debug().Msgf("Skipping zero difference fix for frame %s#%d", id, i)
				fixedFrames = append(fixedFrames, frame)
				continue
			}
			for field, change := range fixes {
				log.Info().Msgf("Fixed frame %s#%d.%s: %s -> %s", id, i, field, change.Old, change.New)
			}
			totalFixedCount += 1
			fixesCount += 1
			fixedFrames = append(fixedFrames, fixedFrame)
		}
		if len(fixedFrames) != len(actualFrames) {
			log.Fatal().Msgf("Number of fixed frames (%d) does not match actual frames count (%d). "+
				"This is probably a bug!", len(fixedFrames), len(actualFrames))
		}
		if fixesCount > 0 {
			tag.DeleteFrames(id)
			for _, t := range fixedFrames {
				tag.AddFrame(id, t)
			}
		}
	}

	if totalErrorsCount > 0 {
		if !forced {
			return fmt.Errorf("got %d error(s) while fixing encoding and aborted", totalErrorsCount)
		}
		log.Error().Msgf("Got %d errors(s) while fixing encoding, proceeding", totalErrorsCount)
	}

	err = tag.Save()
	if err != nil {
		return fmt.Errorf("failed saving temp file: %w", err)
	}
	log.Info().Msgf("Fixed %d frame(s)", totalFixedCount)

	return nil
}

func fixFrame(f id3v2.Framer) (id3v2.Framer, map[string]Change, error) {
	switch v := f.(type) {
	case id3v2.UserDefinedTextFrame:
		val, err := fixEncoding(v.Value)
		if err != nil {
			return nil, nil, err
		}
		if val == v.Value {
			return nil, nil, nil
		}
		v.Value = val
		v.Encoding = id3v2.EncodingUTF8
		return v, map[string]Change{"Value": {v.Value, val}}, nil

	case id3v2.TextFrame:
		text, err := fixEncoding(v.Text)
		if err != nil {
			return nil, nil, err
		}
		if text == v.Text {
			return nil, nil, nil
		}
		v.Text = text
		v.Encoding = id3v2.EncodingUTF8
		return v, map[string]Change{"Text": {v.Text, text}}, nil

	case id3v2.CommentFrame:
		text, err := fixEncoding(v.Text)
		if err != nil {
			return nil, nil, err
		}
		desc, err := fixEncoding(v.Description)
		if err != nil {
			return nil, nil, err
		}
		if text == v.Text && desc == v.Description {
			return nil, nil, nil
		}
		v.Text = text
		v.Description = desc
		v.Encoding = id3v2.EncodingUTF8
		return v, map[string]Change{"Text": {v.Text, text}, "Description": {v.Description, desc}}, nil

	default:
		return nil, nil, errors.New("failed to detect frame type")
	}
}

func supportedMp3Frames() map[string]string {
	supportedFrames := make(map[string]string)
	seenIds := make(map[string]bool)
	for title, id := range id3v2.V23CommonIDs {
		// skip duplicates
		if _, ok := seenIds[id]; ok {
			continue
		}
		if id == "COMM" || id[0] == 'T' {
			supportedFrames[title] = id
			seenIds[id] = true
		}
	}
	return supportedFrames
}
