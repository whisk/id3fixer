# id3fixer

`id3fixer` is a command-line utility designed to correct the encoding issues found in CP1251 (also known as windows-1251 or Cyrillic) MP3 tags. Currently only ID3v1, ID3v2.3, ID3v2.4 are supported.

## Synopsis
```
Usage:
       id3fixer -src <source_file.mp3> [-dst <destination_file.mp3>]
or
       id3fixer <source_file 1.mp3> [<source_file 2.mp3> ...]
Arguments:
  -dst string
    	destination file name. Default: empty (fix in-place)
  -f	be forceful, do not abort on encoding errors
  -frames value
    	comma-separated list of frames to fix (only for id3v2) (default TRSO,TIT3,TPE1,TRDA,TCOP,TIME,COMM,TIT1,TOWN,TXXX,TRCK,TMED,TOAL,TPE3,TDAT,TIT2,TOPE,TLEN,TBPM,TSRC,TEXT,TPE4,TCON,TOLY,TFLT,TPOS,TSSE,TENC,TSIZ,TDLY,TCOM,TYER,TALB,TKEY,TPUB,TLAN,TORY,TOFN,TRSN,TPE2)
  -h	show help message
  -l	show a full list of supported id3v2 frames
  -src string
    	source file name
  -v	be verbose
  -version
    	show version information
  -vv
    	be very verbose (implies -v)
```

## TODO

- [x] add support for reading ID3v1

## License

This project is licensed under the MIT License.
