# manifesto
Manifesto is an HLS tool for creating multiple variants, a master.m3u8 file, and converting 608 captions to segmented webvtt subtitles via ffmpeg.

## ``` Setup ```

#### ```Required``` 
* Go 
* Ffmpeg

#### ```Install go```
      https://golang.org/doc/install

#### ```Set your Environment```
```
mkdir -p ~/go/bin
export GOPATH=~/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```
#### ```Install ffmpeg with libx264 support```


## ``` How It Works ```

Manifesto transcodes and segments video into multiple variants and creates the master.m3u8 file. 
608 Closed captions are extracted and converted to webvtt segment files.

### ``` Quick Start```

* ``` go build manifesto.go ```
* ``` ./manifesto -i vid.ts ```

This will create the following directory structure and files 

```
vid:
hd720  low640  master.m3u8  med960  subs

vid/med960:
index0.ts  index1.ts  index2.ts  index3.ts  index4.ts  index.m3u8

vid/hd720:
index0.ts   index1.ts   index2.ts   index3.ts   index4.ts   index.m3u8

vid/low640:
index0.ts   index1.ts   index2.ts   index3.ts   index4.ts   index.m3u8

vid/subs:
index0.vtt  index1.vtt  index2.vtt  index3.vtt  index4.vtt  index_vtt.m3u8
```

* The default toplevel directory name is the video file name without the file extention.
* The variants are read from the hls.json file, variants can be added or removed as needed. 
* The command used to traanscode is specified in the cmd.template file, it can be modified. 

### ```Command line switches```
```
  -d string
    	override top level directory for hls files (optional)
  -i string
    	Video file to segment (required)
  -j string
    	JSON file of variants (optional) (default "./hls.json")
  -s string
    	subtitle file to segment (optional)
  -t string
    	command template file (optional) (default "./cmd.template")

```






