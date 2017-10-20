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
	 . "github.com/logrusorgru/aurora"

)

var infile string
var subfile string
var toplevel string

var jasonfile string
var cmdtemplate string
var captioned bool
var completed string
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
	if captioned {
		inputs = mkCaptionInputs(infile)
	}
	if hasSideCar() {
		inputs = mkSubInputs(infile, subfile)
	}
	r := strings.NewReplacer("INFILE", inputs, "ASPECT", v.Aspect,
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
	fmt.Printf("* Variants: %s %s \r", Green(completed),v.Aspect)
	completed +=fmt.Sprintf("%s ",v.Aspect)
	cmd := v.mkCmd(cmdtemplate)
	chkExec(cmd)
	v.readRate()
	if hasCapsOrSubs() {
		srcdir := fmt.Sprintf("%s/%s", toplevel, v.Name)
		mvCaptions(srcdir)
		captioned = false
		subfile = ""
	}
	fmt.Printf("* Variants: %s  \r", Green(completed))
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	stanza := fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v", v.Bandwidth, v.Aspect)
	if hasCapsOrSubs() {
		stanza =fmt.Sprintf("%s,SUBTITLES=\"webvtt\"",stanza)
		}
		return stanza
}

func hasSideCar()bool{
	if subfile !="" {
		return true
	}
	return false	
}
func hasCapsOrSubs()bool{
	if (captioned) || hasSideCar() {
		return true
		}
		return false
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
		}
	}
}

// return a string for file inputs in the ffmpeg command to extract 608 captions for vtt subtitles
func mkCaptionInputs(infile string) string {
	return fmt.Sprintf("%s -f lavfi -fix_sub_duration -i movie=%s[out0+subcc] ", infile, infile)
}

//return a string for file inputs in the ffmpeg command to use an external subtitle file for vtt subtitles
func mkSubInputs(infile string, subfile string) string {
	return fmt.Sprintf("%s -fix_sub_duration -i %s ", infile, subfile)
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
	if !(hasSideCar()) {
		chkCaptions(infile)
	}
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", toplevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	if hasCapsOrSubs(){
		w.WriteString(mkSubStanza())
	}	
	fmt.Println("\n* Video file:",Cyan(infile),"\n* Toplevel:",Cyan(toplevel),"\n* Subtitle file:",Cyan(subfile),"\n* Captions:",Cyan(captioned) ) 
	
	for _, v := range variants {
		v.start()
		w.WriteString(fmt.Sprintf("%s\n", v.mkStanza()))
		w.WriteString(fmt.Sprintf("%s/index.m3u8\n", v.Name))
	}
	fmt.Println("\n\n")
	w.Flush()
}

func main() {
	flag.StringVar(&infile, "i", "", "Video file to segment (required)")
	flag.StringVar(&subfile, "s", "", "subtitle file to segment (optional)")
	flag.StringVar(&toplevel, "d", "", "override top level directory for hls files (optional)")
	flag.StringVar(&jasonfile, "j", `./hls.json`, "JSON file of variants (optional)")
	flag.StringVar(&cmdtemplate, "t", `./cmd.template`, "command template file (optional)")

	flag.Parse()
	variants := dataToVariants()

	if infile != "" {
		mkAll(variants)
	} else {
		flag.PrintDefaults()
	}
}

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
	dest := v.mkDest()
	fmt.Println("\tStarting ", dest,"variant")
	cmd := v.mkCmd(cmdtemplate)
	chkExec(cmd)
	v.readRate()
	if hasCapsOrSubs() {
		srcdir := fmt.Sprintf("%s/%s", toplevel, v.Name)
		mvCaptions(srcdir)
		captioned = false
		subfile = ""
	}
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	stanza := fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v", v.Bandwidth, v.Aspect)
	if hasCapsOrSubs() {
		stanza =fmt.Sprintf("%s,SUBTITLES=\"webvtt\"",stanza)
		}
		return stanza
}

func hasCapsOrSubs()bool{
	if (captioned) || (subfile != "") {
		return true
		}
		return false
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
			fmt.Println("\t\tMoving", f.Name(), "to subs dir")
		}
	}
}

// return a string for file inputs in the ffmpeg command to extract 608 captions for vtt subtitles
func mkCaptionInputs(infile string) string {
	return fmt.Sprintf("%s -f lavfi -fix_sub_duration -i movie=%s[out0+subcc] ", infile, infile)
}

//return a string for file inputs in the ffmpeg command to use an external subtitle file for vtt subtitles
func mkSubInputs(infile string, subfile string) string {
	return fmt.Sprintf("%s -fix_sub_duration -i %s ", infile, subfile)
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
	if subfile == "" {
		chkCaptions(infile)
	}
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", toplevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	if hasCapsOrSubs(){
		w.WriteString(mkSubStanza())
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
	flag.StringVar(&subfile, "s", "", "subtitle file to segment (optional)")
	flag.StringVar(&toplevel, "d", "", "override top level directory for hls files (optional)")
	flag.StringVar(&jasonfile, "j", `./hls.json`, "JSON file of variants (optional)")
	flag.StringVar(&cmdtemplate, "t", `./cmd.template`, "command template file (optional)")

	flag.Parse()
	variants := dataToVariants()

	if infile != "" {
		if toplevel == "" {
			toplevel = setTop()
		}
		fmt.Println("\nTop level set to", toplevel)

		mkAll(variants)
	} else {
		flag.PrintDefaults()
	}
}
