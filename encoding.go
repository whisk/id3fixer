package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
)

// correctly re-encode 1 byte cp1251 string to a valid utf8
func fixEncoding(s string) (string, error) {
	log.Trace().Msg("Fix bytes: \n" + formatBytes([]byte(s)))

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

func formatBytes(arr []byte) string {
	s := ""
	for _, b := range arr {
		s += fmt.Sprintf("%08b ", b)
	}
	return s
}
