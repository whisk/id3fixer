package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type framesMap map[string]string

type optionsType struct {
	src        string
	sources    []string
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
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = time.DateTime
	log.Logger = zerolog.New(consoleWriter).With().Timestamp().Logger()

	if options.listFrames {
		fmt.Println("Suported frames:")
		for title, id := range supportedMp3Frames() {
			fmt.Printf("%s\t%s\n", id, title)
		}
		os.Exit(0)
	} else if options.help || (options.src == "" && len(options.sources) == 0) || (len(options.sources) > 0 && options.dst != "") {
		progname := filepath.Base(os.Args[0])
		fmt.Printf("Usage:\n")
		fmt.Printf("       %s -src <source_file.mp3> [-dst <destination_file.mp3>]\n", progname)
		fmt.Printf("       %s <source_file 1.mp3> [<source_file 2.mp3> ...]\n", progname)
		fmt.Println("Arguments:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	errCnt := 0
	if len(options.sources) > 0 {
		fixedCnt := 0
		for _, src := range options.sources {
			log.Info().Msgf("Fixing %s...", src)
			err := fixMp3(src, "", options.frames, options.forced)
			if err != nil {
				log.Error().Err(err).Msg("")
				// for debug purposes
				if unwrapped := errors.Unwrap(err); unwrapped != nil {
					log.Error().Err(unwrapped).Msg("Unwrapped error")
				}
				errCnt += 1
				if !options.forced {
					log.Error().Msg("Aborting...")
					break
				}
			} else {
				fixedCnt += 1
			}
		}
		log.Info().Msgf("Fixed %d/%d files", fixedCnt, len(options.sources))
	} else {
		err := fixMp3(options.src, options.dst, options.frames, options.forced)
		if err != nil {
			log.Error().Err(err).Msg("")
			// for debug purposes
			if unwrapped := errors.Unwrap(err); unwrapped != nil {
				log.Error().Err(unwrapped).Msg("Unwrapped error")
			}
			errCnt += 1
		}
	}
	if errCnt > 0 {
		os.Exit(1)
	}
}

func parseCmdlineOptions() optionsType {
	options := optionsType{}
	flag.StringVar(&options.src, "src", "", "source file name")
	flag.StringVar(&options.dst, "dst", "", "destination file name. Default: empty (fix in-place)")
	flag.Var(&options.frames, "frames", "comma-separated list of frames to fix")
	flag.BoolVar(&options.listFrames, "l", false, "show a full list of supported frames")
	flag.BoolVar(&options.forced, "f", false, "be forceful, do not abort on encoding errors")
	flag.BoolVar(&options.verbose, "v", false, "be verbose")
	flag.BoolVar(&options.vverbose, "vv", false, "be very verbose (implies -v)")
	flag.BoolVar(&options.help, "h", false, "show help message")

	flag.Parse()
	options.sources = flag.Args()

	return options
}
