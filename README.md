# id3fixer

`id3fixer` is a command-line utility designed to correct the encoding issues found in CP1251 (also known as windows-1251 or Cyrillic) MP3 tags. Currently only ID3v2.3 and ID3v2.4 are supported for reading, the fixed output is always ID3v2.4.

## Synopsis
```
Usage: id3fixer -src <source_file.mp3> [-dst <destination_file.mp3>]
Arguments:
  -dst string
    	destination file name. Default: empty (fix in-place)
  -f	be forceful, do not abort on encoding errors (default true)
  -frames value
    	comma-separated list of frames to fix (default TPOS,TOWN,TRCK,TPE1,COMM,TMED,TIT3,TPE2,TIT1,TRDA,TRSO,TIME,TPE4,TALB,TOLY,TCOP,TPE3,TSSE,TLEN,TORY,TSIZ,TBPM,TYER,TCOM,TENC,TDLY,TOPE,TXXX,TIT2,TEXT,TDAT,TOFN,TCON,TKEY,TRSN,TPUB,TSRC,TOAL,TLAN,TFLT)
  -h	show help message
  -l	show a full list of supported frames
  -src string
    	source file name
  -v	be verbose
  -vv
    	be very verbose (implies -v)
```

## TODO

* add support for reading ID3v1

## License

This project is licensed under the MIT License.