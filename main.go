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
// + write readme
// - add golden test

type framesMap map[string]string

type optionsType struct {
	src        string
	dst        string
	frames     framesMap
	listFrames bool
	forced     bool
	verbose    bool
	vverbose   bool
	help       bool
}

// sets frames to fix cmdline option
func (f *framesMap) Set(value string) error {
	rawFrames := strings.Split(value, ",")
	supportedFrames := supportedMp3Frames()
	if len(rawFrames) == 0 || rawFrames[0] == "ALL" {
		log.Debug().Msg("Fixing all supported frames")
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
	options := parseCmdlineOptions()

	if options.vverbose {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if options.verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	if options.listFrames {
		fmt.Println("Suported frames:")
		for title, id := range supportedMp3Frames() {
			fmt.Printf("%s\t%s\n", id, title)
		}
		os.Exit(0)
	} else if options.help || options.src == "" {
		fmt.Printf("Usage: %s -src <source_file.mp3> [-dst <destination_file.mp3>]\n", os.Args[0])
		fmt.Println("Arguments:")
		flag.PrintDefaults()
		os.Exit(1)
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
	flag.StringVar(&options.dst, "dst", "", "destination file name. Default: fix in-place")
	flag.Var(&options.frames, "frames", "comma-separated list of frames to fix")
	flag.BoolVar(&options.listFrames, "l", false, "show a full list of supported frames")
	flag.BoolVar(&options.forced, "f", true, "be forceful, do not abort on encoding errors")
	flag.BoolVar(&options.verbose, "v", false, "be verbose")
	flag.BoolVar(&options.vverbose, "vv", false, "be very verbose (implies -v)")
	flag.BoolVar(&options.help, "h", false, "show help message")

	flag.Parse()

	return options
}
