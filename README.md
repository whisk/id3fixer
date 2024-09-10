# id3fixer

`id3fixer` is a command-line utility designed to correct the encoding issues found in CP1251 (also known as windows-1251 or Cyrillic) MP3 tags. Currently only ID3v2.3 and ID3v2.4 are supported for reading, the fixed output is always ID3v2.4.

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
    	comma-separated list of frames to fix (default TIT1,TPE4,TPOS,TRSO,TPE1,TPE3,TSRC,TSSE,COMM,TIT2,TOWN,TOAL,TALB,TENC,TOFN,TXXX,TCOP,TOPE,TIT3,TCOM,TRCK,TLAN,TPUB,TDLY,TORY,TRSN,TKEY,TEXT,TOLY,TYER,TCON,TLEN,TFLT,TMED,TSIZ,TPE2,TIME,TDAT,TBPM,TRDA)
  -h	show help message
  -l	show a full list of supported frames
  -src string
    	source file name
  -v	be verbose
  -version
    	show version information
  -vv
    	be very verbose (implies -v)
```

## TODO

* add support for reading ID3v1

## License

This project is licensed under the MIT License.
