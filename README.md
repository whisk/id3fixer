# id3fixer

`id3fixer` is a command-line utility designed to correct the encoding issues found in CP1251 (also known as windows-1251 or Cyrillic) MP3 tags. Currently only ID3v2.3 and ID3v2.4 are supported.

## Synopsis
```
Usage: id3fixer -src <source_file.mp3> [-dst <destination_file.mp3>]
Arguments:
  -dst string
    	destination file name. Default: fix in-place
  -f	be forceful, do not abort on encoding errors (default true)
  -frames value
    	comma-separated list of frames to fix (default TPE2,TSRC,TFLT,TPE1,TIT3,TIT1,TEXT,TOPE,TSIZ,TCOP,TIT2,TCOM,TPE4,TBPM,TPE1,TLAN,TLEN,TIT2,TOWN,TRCK,TALB,TMED,COMM,TIME,TOAL,TKEY,TPOS,TORY,TCON,TXXX,TPE3,TENC,TOLY,TRSN,TPUB,TCON,TOFN,TYER,TRSO,TDAT,TSSE,TDLY,TRDA)
  -h	show help message
  -l	show a full list of supported frames
  -src string
    	source file name
  -v	be verbose
  -vv
    	be very verbose (implies -v)
```

## License

This project is licensed under the MIT License.