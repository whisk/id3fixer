package main

import (
	"errors"
	"flag"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TODO
// + file copy
// + temp file creation
// + logging/debug levels
// + cmdline options
// + push to github
// + modularity
// + error wrapping/unwrapping
// + fix in-place (with backups)

type optionsType struct {
	src       string
	dst       string
	fixTitle  bool
	fixArtist bool
	fixAlbum  bool
}

func main() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	options := parseCmdlineOptions()

	err := fixMp3(options.src, options.dst, options.fixTitle, options.fixArtist, options.fixAlbum)
	if err != nil {
		log.Error().Err(err).Msg("")
		// for debug purposes
		if unwrapped := errors.Unwrap(err); unwrapped != nil {
			log.Error().Err(unwrapped).Msg("Unwrapped error")
		}
		os.Exit(1)
	}
}

func parseCmdlineOptions() optionsType {
	options := optionsType{}
	flag.StringVar(&options.src, "src", "", "source file name")
	flag.StringVar(&options.dst, "dst", "", "destination file name")
	flag.BoolVar(&options.fixTitle, "fix-title", true, "fix title")
	flag.BoolVar(&options.fixArtist, "fix-artist", true, "fix artist")
	flag.BoolVar(&options.fixAlbum, "fix-album", true, "fix album")

	flag.Parse()

	return options
}
