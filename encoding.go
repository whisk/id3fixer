package main

import (
	"fmt"
	"unicode/utf8"

	translit "github.com/essentialkaos/translit/v3"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
)

// correctly re-encodes 1-byte cp1251, erroneously encoded as a 2-byte string, back to a valid utf8
func fixEncoding(s string) (string, error) {
	log.Trace().Msg("Fix bytes (bin): \n" + formatBytes([]byte(s)))
	log.Trace().Msg("Fix bytes (dec): \n" + formatBytes10([]byte(s)))

	// this removes invalid utf8 "first" bytes from byte pairs, leaving only "second" bytes with correct padding
	// Windows1252 is an arbitrary 1 byte encoder, as it supports more runes than other 125X
	// output string is a correct 1 byte cp1251 string
	encoder := charmap.Windows1252.NewEncoder()
	res, err := encoder.String(s)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("UTF8->1byte: \n" + formatBytes([]byte(res)))

	// now encode it into utf8
	decoder := charmap.Windows1251.NewDecoder()
	res, err = decoder.String(res)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("1byte->UTF8: \n" + formatBytes([]byte(res)))

	return res, nil
}

// fixEncodingCp1251 converts 1-byte cp1251 string to a valid latin1
func fixEncodingCp1251(s string, maxByteLength int) (string, error) {
	log.Trace().Msg("Fix bytes (bin): \n" + formatBytes([]byte(s)))
	log.Trace().Msg("Fix bytes (dec): \n" + formatBytes10([]byte(s)))

	decoder := charmap.Windows1251.NewDecoder()
	res, err := decoder.String(s)
	if err != nil {
		return "", err
	}
	log.Trace().Msg("1byte->UTF8: \n" + formatBytes([]byte(res)))

	transliterator := translit.ICAO
	translit := transliterator(res)
	log.Trace().Msg("UTF8->translit: \n" + translit)

	return truncateBytes(translit, maxByteLength), nil
}

// truncateBytes truncates a string to a maximum byte length, ensuring that it does not cut off
// a UTF-8 character in the middle
func truncateBytes(s string, maxByteLength int) string {
	limit := maxByteLength
	if len(s) <= limit {
		return s
	}
	for limit > 0 && !utf8.RuneStart(s[limit]) {
		limit -= 1
	}
	s = s[:limit]
	log.Debug().Msgf("truncated string to %d (<=%d) bytes: %s", limit, maxByteLength, s)

	return s
}

func formatBytes(arr []byte) string {
	s := ""
	for _, b := range arr {
		s += fmt.Sprintf("%08b ", b)
	}
	return s
}

func formatBytes10(arr []byte) string {
	s := ""
	for _, b := range arr {
		s += fmt.Sprintf("%08d ", b)
	}
	return s
}
