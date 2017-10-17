package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var infile string
var subfile string
var toplevel string

var jasonfile string
var cmdtemplate string
var captioned bool

// Variant struct for HLS variants
type Variant struct {
	Name      string `json:"name"`
	Aspect    string `json:"aspect"`
	Rate      string `json:"framerate"`
	Vbr       string `json:"vbitrate"`
	Abr       string `json:"abitrate"`
	Bandwidth string
}

// Create variant's destination directory
func (v *Variant) mkDest() string {
	dest := fmt.Sprintf("%s/%s", toplevel, v.Name)
	os.MkdirAll(dest, 0755)
	return dest
}

// This Variant method assembles the ffmpeg command
func (v *Variant) mkCmd(cmdtemplate string) string {
	data, err := ioutil.ReadFile(cmdtemplate)
	chk(err, "Error reading template file")
	inputs := infile
	r := strings.NewReplacer("INFILE", inputs, "ASPECT", v.Aspect,
		"VBITRATE", v.Vbr, "FRAMERATE", v.Rate, "ABITRATE", v.Abr,
		"TOPLEVEL", toplevel, "NAME", v.Name, "\n", " ")
	cmd := fmt.Sprintf("%s\n", r.Replace(string(data)))

	return cmd
}

// Read actual bitrate from first segment to set bandwidth in master.m3u8
func (v *Variant) readRate() {
	cmd := fmt.Sprintf("ffprobe -i %s/%s/index0.ts", toplevel, v.Name)
	data := chkExec(cmd)
	two := strings.Split(data, "bitrate: ")[1]
	rate := strings.Split(two, " kb/s")[0]
	v.Bandwidth = fmt.Sprintf("%v000", rate)
}

// Start transcoding the variant
func (v *Variant) start() {
	dest := v.mkDest()
	fmt.Println("Starting ", dest)
	cmd := v.mkCmd(cmdtemplate)
	chkExec(cmd)
	v.readRate()
	if captioned {
		srcdir := fmt.Sprintf("%s/%s", toplevel, v.Name)
		mvCaptions(srcdir)
		captioned = false
	}
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	return fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v,SUBTITLES=\"webvtt\"", v.Bandwidth, v.Aspect)

}

func chkExec(cmd string) string {
	// Executes external commands and checks for runtime errors
	parts := strings.Fields(cmd)
	data, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	chk(err, fmt.Sprintf("Error running \n %s \n %v", cmd, string(data)))
	return string(data)
}

// probes for Closed Captions in video file.
func chkCaptions(mediafile string) {
	captioned = false
	cmd := fmt.Sprintf("ffprobe -i %s", mediafile)
	data := chkExec(cmd)
	if strings.Contains(data, "Captions") {
		fmt.Println("608 captions detected")
		captioned = true
	}

}

// Captions are segmented along with the first variant and then moved to toplevel/subs
func mvCaptions(srcdir string) {
	destdir := fmt.Sprintf("%s/subs", toplevel)
	os.MkdirAll(destdir, 0755)
	files, err := ioutil.ReadDir(srcdir)
	chk(err, "Error moving Captions")
	for _, f := range files {
		if strings.Contains(f.Name(), "vtt") {
			os.Rename(fmt.Sprintf("%s/%s", srcdir, f.Name()), fmt.Sprintf("%s/%s", destdir, f.Name()))
			fmt.Println(f.Name())
		}
	}
}

// Read json file for variants
func dataToVariants() []Variant {
	var variants []Variant
	data, err := ioutil.ReadFile(jasonfile)
	chk(err, "Error reading JSON file")
	json.Unmarshal(data, &variants)
	return variants
}

// Set the toplevel dir for variants by splitting video file name at the "."
func setTop() string {
	return strings.Split(infile, `.`)[0]

}

// Generic catchall error checking
func chk(err error, mesg string) {
	if err != nil {
		fmt.Printf("%s\n", mesg)
		//panic(err)
	}
}

// Make all variants and write master.m3u8
func mkAll(variants []Variant) {
	os.MkdirAll(toplevel, 0755)
	chkCaptions(infile)
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", toplevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	if captioned {
		w.WriteString("#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"webvtt\",NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,LANGUAGE=\"en\",URI=\"subs/vtt_index.m3u8\"\n")
	}
	for _, v := range variants {
		v.start()
		w.WriteString(fmt.Sprintf("%s\n", v.mkStanza()))
		w.WriteString(fmt.Sprintf("%s/index.m3u8\n", v.Name))
	}
	w.Flush()
}

func main() {
	flag.StringVar(&infile, "i", "", "Video file to segment (required)")
	flag.StringVar(&toplevel, "d", "", "override top level directory for hls files")
	flag.StringVar(&jasonfile, "j", `./hls.json`, "JSON file of variants")
	flag.StringVar(&cmdtemplate, "t", `./cmd.template`, "command template file")

	flag.Parse()
	variants := dataToVariants()

	if infile != "" {
		if toplevel == "" {
			toplevel = setTop()
		}
		fmt.Println("Top level set to ", toplevel)

		mkAll(variants)
	} else {
		flag.PrintDefaults()
	}
}
