package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

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
// + support fixing comments
// - write readme

type framesMap map[string]string

type optionsType struct {
	src        string
	dst        string
	frames     framesMap
	listFrames bool
	forced     bool
}

// sets frames to fix cmdline option
func (f *framesMap) Set(value string) error {
	rawFrames := strings.Split(value, ",")
	supportedFrames := supportedMp3Frames()
	if len(rawFrames) == 0 || rawFrames[0] == "ALL" {
		log.Trace().Msg("No frames to fix given, falling back to all supported frames")
		*f = supportedFrames
		return nil
	}
	setFrames := make(map[string]string)
	for i := range rawFrames {
		// trim for no reason
		frameId := strings.TrimSpace(rawFrames[i])
		found := false
		for title, id := range supportedFrames {
			// inefficient, but ok
			if frameId == id {
				setFrames[title] = id
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("frame %s not supported", frameId)
		}
	}
	*f = setFrames
	return nil
}

// reads frames to fix cmdline option as a string
func (f *framesMap) String() string {
	if len(*f) == 0 {
		// we set default value here
		*f = supportedMp3Frames()
	}
	t := make([]string, 0, len(*f))
	for _, id := range *f {
		t = append(t, id)
	}
	return strings.Join(t, ",")
}

func main() {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	options := parseCmdlineOptions()
	if options.listFrames {
		for title, id := range supportedMp3Frames() {
			fmt.Printf("%s\t%s\n", id, title)
		}
		os.Exit(0)
	}

	err := fixMp3(options.src, options.dst, options.frames, options.forced)
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
	flag.Var(&options.frames, "frames", "comma-separated list of frames to fix. Default: ALL")
	flag.BoolVar(&options.listFrames, "l", false, "show list of supported frames")
	flag.BoolVar(&options.forced, "f", true, "be forceful, do not stop on encoding errors")

	flag.Parse()

	return options
}
