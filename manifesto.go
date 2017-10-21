package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/logrusorgru/aurora"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var infile string
var subfile string
var addsubs bool
var toplevel string
var webvtt bool
var jasonfile string
var cmdtemplate string
var completed string
var batch string

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

func (v *Variant) mkInputs() string {
	inputs := fmt.Sprintf(" -i %s", infile)
	if addsubs {
		inputs = fmt.Sprintf(" -i %s -i %s  ", infile, subfile)
	}
	return inputs
}

// This Variant method assembles the ffmpeg command
func (v *Variant) mkCmd(cmdtemplate string) string {
	data, err := ioutil.ReadFile(cmdtemplate)
	chk(err, "Error reading template file")
	inputs := v.mkInputs()
	r := strings.NewReplacer("INPUTS", inputs, "ASPECT", v.Aspect,
		"VBITRATE", v.Vbr, "FRAMERATE", v.Rate, "ABITRATE", v.Abr,
		"TOPLEVEL", toplevel, "NAME", v.Name, "\n", " ")
	cmd := fmt.Sprintf("%s\n", r.Replace(string(data)))
	//fmt.Println(cmd)
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
	v.mkDest()
	fmt.Printf("* Variants: %s %s \r", Cyan(completed), v.Aspect)
	completed += fmt.Sprintf("%s ", v.Aspect)
	cmd := v.mkCmd(cmdtemplate)
	chkExec(cmd)
	v.readRate()
	fmt.Printf("* Variants: %s  \r", Cyan(completed))
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	stanza := fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v", v.Bandwidth, v.Aspect)
	if addsubs {
		stanza = fmt.Sprintf("%s,SUBTITLES=\"webvtt\"", stanza)
	}
	return stanza
}

func chkExec(cmd string) string {
	// Executes external commands and checks for runtime errors
	parts := strings.Fields(cmd)
	data, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	chk(err, fmt.Sprintf("Error running \n %s \n %v", cmd, string(data)))
	return string(data)
}

// probes for Closed Captions in video file.
func hasCaptions() bool {
	cmd := fmt.Sprintf("ffprobe -i %s", infile)
	data := chkExec(cmd)
	if strings.Contains(data, "Captions") {
		return true
	}
	return false
}

// Captions are segmented along with the first variant and then moved to toplevel/subs
func mvCaptions(vardir string) {
	srcdir := fmt.Sprintf("%s/%s", toplevel, vardir)
	destdir := fmt.Sprintf("%s/subs", toplevel)
	os.MkdirAll(destdir, 0755)
	files, err := ioutil.ReadDir(srcdir)
	chk(err, "Error moving Captions")
	for _, f := range files {
		if strings.Contains(f.Name(), "vtt") {
			os.Rename(fmt.Sprintf("%s/%s", srcdir, f.Name()), fmt.Sprintf("%s/%s", destdir, f.Name()))
		}
	}
}

// return a subtitle stanza for use in the  master.m3u8
func mkSubStanza() string {
	return "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"webvtt\",NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,LANGUAGE=\"en\",URI=\"subs/vtt_index.m3u8\"\n"
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
func mkTopLevel() {
	if toplevel == "" {
		toplevel = strings.Split(infile, `.`)[0]
	}
	os.MkdirAll(toplevel, 0755)
}

func extractCaptions() string {
	fmt.Println("Extracting captions")
	assfile := fmt.Sprintf("%s/%s.ass",toplevel, toplevel)
	cmd := fmt.Sprintf("ffmpeg -y -f lavfi -fix_sub_duration -i movie=%s[out0+subcc] %s", infile, assfile)
	chkExec(cmd)
	return assfile
}

func mkSubfile() {
	if webvtt {
		return
	} else {
		addsubs = false
		if subfile == "" {
			if hasCaptions() {
				subfile = extractCaptions()
			}
		}
		if subfile != "" {
		
			addsubs = true
		}
	}
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
	mkTopLevel()
	mkSubfile()
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", toplevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	fmt.Println("\n* Video file:", Cyan(infile), "\n* Toplevel:", Cyan(toplevel), "\n* Subtitle file:", Cyan(subfile))
	for _, v := range variants {
		v.start()
		if addsubs && !(webvtt) {
			mvCaptions(v.Name)
			w.WriteString(mkSubStanza())
			addsubs = false
			subfile = ""
			webvtt = true
		}
		w.WriteString(fmt.Sprintf("%s\n", v.mkStanza()))
		w.WriteString(fmt.Sprintf("%s/index.m3u8\n", v.Name))
	}
	fmt.Println("\n\n")
	w.Flush()
}

func main() {
	flag.StringVar(&infile, "i", "", "Video file to segment (either -i or -b is required)")
	flag.StringVar(&subfile, "s", "", "subtitle file to segment (optional)")
	flag.StringVar(&toplevel, "d", "", "override top level directory for hls files (optional)")
	flag.StringVar(&jasonfile, "j", `./hls.json`, "JSON file of variants (optional)")
	flag.StringVar(&cmdtemplate, "t", `./cmd.template`, "command template file (optional)")
	flag.StringVar(&batch, "b", "", "batch mode, list multiple input files (either -i or -b is required)")

	flag.Parse()
	fmt.Println(batch)
	variants := dataToVariants()

	if batch != "" {
		batch = strings.Replace(batch, " ", ",", -1)
		for _, b := range strings.Split(batch, ",") {
			fmt.Println(b)
			webvtt = false
			infile = b
			completed = ""
			toplevel = ""
			mkTopLevel()
			variants := dataToVariants()
			mkAll(variants)
		}
	} else {
		if infile != "" {
			mkAll(variants)
		} else {
			flag.PrintDefaults()
		}

	}
}
