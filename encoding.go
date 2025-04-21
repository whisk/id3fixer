package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	translit "github.com/essentialkaos/translit/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
)

// brokenCp1251ToUtf8 re-encodes 1-byte cp1251, erroneously encoded as a 2-byte string, back to a valid utf8
func brokenCp1251ToUtf8(s string) (string, error) {
	log.Trace().Msg("Fix bytes (bin):\n" + formatBytes(s))
	log.Trace().Msg("Fix bytes (dec):\n" + formatBytes10(s))

	// this removes invalid utf8 "first" bytes from byte pairs, leaving only "second" bytes with correct padding
	// Windows1252 is an arbitrary 1 byte encoder, as it supports more runes than other 125X
	// output string is a correct 1 byte cp1251 string
	encoder := charmap.Windows1252.NewEncoder()
	res, err := encoder.String(s)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("utf8->1byte:\n" + formatBytes(res))

	// now encode it into utf8
	decoder := charmap.Windows1251.NewDecoder()
	res, err = decoder.String(res)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("1byte->utf8:\n" + formatBytes(res))

	return res, nil
}

// cp1251ToTranslit transliterates a valid 1-byte cp1251 string to a valid latin1
func cp1251ToTranslit(s string, maxByteLength int) (string, error) {
	log.Trace().Msg("Fix bytes (bin):\n" + formatBytes(s))
	log.Trace().Msg("Fix bytes (dec):\n" + formatBytes10(s))

	decoder := charmap.Windows1251.NewDecoder()
	res, err := decoder.String(s)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("1byte->utf8:\n" + formatBytes(s))

	transliterator := translit.ICAO
	translit := transliterator(res)
	log.Trace().Msg("utf8->translit:\n" + translit)

	return truncateUtf8(translit, maxByteLength), nil
}

// truncateUtf8 truncates a string to a maximum byte length, ensuring that it does not cut off
// a UTF-8 character in the middle
func truncateUtf8(s string, maxByteLength int) string {
	limit := maxByteLength
	if len(s) <= limit {
		return s
	}
	for limit > 0 && !utf8.RuneStart(s[limit]) {
		limit -= 1
	}
	s = s[:limit]
	log.Trace().Msgf("truncated string to %d (<=%d) bytes: %s", limit, maxByteLength, s)

	return s
}

func formatBytes(s string) string {
	res := strings.Builder{}
	for i := 0; i < len(s); i++ {
		res.WriteString(fmt.Sprintf("%08b ", s[i]))
	}
	return res.String()
}

func formatBytes10(s string) string {
	res := strings.Builder{}
	for i := 0; i < len(s); i++ {
		res.WriteString(fmt.Sprintf("%08d ", s[i]))
	}
	return res.String()
}
