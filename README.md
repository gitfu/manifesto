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
* ``` ./manifesto -i video.file ```



