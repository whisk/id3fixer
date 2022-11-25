package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"golang.org/x/text/encoding/charmap"
)

func fixEncoding(s string) (string, error) {
	log.Debug().Msg("Bytes: \n" + formatBytes([]byte(s)))

	encoder := charmap.Windows1252.NewEncoder()
	res, err := encoder.String(s)
	if err != nil {
		return "", err
	}
	log.Debug().Msg("1 pass: \n" + formatBytes([]byte(res)))

	decoder := charmap.Windows1251.NewDecoder()
	res, err = decoder.String(res)
	if err != nil {
		return "", err
	}
	log.Debug().Msg("2 pass: \n" + formatBytes([]byte(res)))

	return res, nil
}

func formatBytes(arr []byte) string {
	s := ""
	for _, b := range arr {
		s += fmt.Sprintf("%08b ", b)
	}
	return s
}
