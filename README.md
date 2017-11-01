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


#### ```Git clone the repo ```
```
git clone https://github.com/gitfu/manifesto
cd manifesto
go build manifesto.go
```

## ``` How It Works ```

Manifesto transcodes and segments video into multiple variants and creates the master.m3u8 file. 
608 Closed captions are extracted and converted to webvtt segment files.
Bandwidth values are automatically calculated to be accurate.

### ``` Quick Start```

* ``` cd ~/manifesto ```
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
### ``` Batch Mode ```
*  the video list has to be either be quoted or comma seperated

```

leroy@futronic:~/manifesto$ ./manifesto -b one.ts,two.ts,three.ts,four.ts,five.ts

 1 of 5
 . Oct 22 20:17:00
 . video file   : one.ts
 . toplevel dir : one
 . caption file : one.ts 
 . subtitle file: one/one.vtt
 . variant sizes: 960x540 768x432 640x360 480x270 1280x720  
 2 of 5
 . Oct 22 20:19:28
 . video file   : two.ts
 . toplevel dir : two
 . caption file : two.ts 
 . subtitle file: two/two.vtt
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
  -u string
    	url prefix to add to index.m3u8 path in master.m3u8 (optional)

```

### ``` /usage ```

```
./manifesto -i vid.mp4
```

* This is single mode a master.m3u8 and variants will be created in a new directory named vid. It will also attempt to extract 608 captions and convert them to segmented webvtt subtitles. 
```
./manifesto -i vid.mp4 -s sub.srt
```
* As above but instead of extracting 608 captions, sub.srt will be converted to a webvtt file and then segmented.

```
./manifesto -i vid.mp4 -s sub.srt -u http://example.com
```
* As above and also adds the url prefix to each variant listed in the m3u8 file. 

```
./manifesto -b a.mov,b.mp4,c.ts,d.mpg 
```
* This is batch mode and it will create directories, a master.m3u8 and variants for each of the files listed. It will also attempt to extracted 608 captions and convert them to segmented webvtt subtitles for each of them. 
```
./manifesto -b a.mov,b.mp4,c.ts,d.mpg -u http://example.com
```
* Does as above but adds the url prefix to each of the variants listed in the m3u8.



### ``` Variants ```


*     Variant data is stored in the hls.json file. 
*     Add or edit or remove as desired.

```
[
{"name": "med960", "aspect": "960x540", "framerate":"29.97","vbitrate": "2000","abitrate": "96k"}
,{"name": "med768", "aspect": "768x432", "framerate":"29.97","vbitrate": "1100","abitrate": "96k"}
,{"name": "low640", "aspect": "640x360", "framerate":"29.97","vbitrate": "730","abitrate": "64k"}
,{"name": "low480", "aspect": "480x270", "framerate":"15","vbitrate": "365","abitrate": "64k"}
,{"name":"hd720","aspect": "1280x720", "framerate" :"29.97","vbitrate": "4500","abitrate": "128k"}

]
```







