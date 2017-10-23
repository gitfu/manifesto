[![Go Report Card](https://goreportcard.com/badge/github.com/gitfu/manifesto)](https://goreportcard.com/report/github.com/gitfu/manifesto)

# Manifesto
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
#### ``` Add one library ```
```
go get -u github.com/logrusorgru/aurora
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

## OR 
### ``` Batch Mode (Hell Yes)```
* ```./manifesto -b vid.ts,vidtwo,ts fu.ts'  ```
### ```the video list has to be either be quoted or comma seperated``` 

```

leroy@futronic:~/scratch/manifesto$ ./manifesto -b one.ts,two.ts,three.ts,four.ts,five.ts

 1 of 5
 . Oct 22 20:17:00
 . video file   : one.ts
 . toplevel dir : one
 . caption file : one.ts 
 . subtitle file: one/one.vtt
 . variant sizes: 960x540 768x432 640x360 480x270 1280x720  

```

* The default toplevel directory name is the video file name without the file extention.
* The variants are read from the hls.json file, variants can be added or removed as needed. 
* The command used to traanscode is specified in the cmd.template file, it can be modified. 

### ```Command line switches```
```
   -b string
    	batch mode, list multiple input files (either -i or -b is required)
  -d string
    	override top level directory for hls files (optional)
  -i string
    	Video file to segment (either -i or -b is required)
  -j string
    	JSON file of variants (optional) (default "./hls.json")
  -s string
    	subtitle file to segment (optional)
  -t string
    	command template file (optional) (default "./cmd.template")

```






