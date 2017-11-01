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
	"time"
)

var infile string
var subfile string
var addsubs bool
var toplevel string
var webvtt bool
var jasonfile string
var cmdtemplate string
var incomplete string
var completed string
var batch string
var urlprefix string
var x264level = "3.0"
var x264profile = "high"
var mastercodec = "avc1.64001E,mp4a.40.2"

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
	if addsubs && !(webvtt) {
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
		"VBITRATE", v.Vbr, "X264LEVEL", x264level,
		"X264PROFILE", x264profile, "FRAMERATE", v.Rate,
		"ABITRATE", v.Abr, "TOPLEVEL", toplevel,
		"NAME", v.Name, "\n", " ")
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
	v.mkDest()
	fmt.Printf(" . variant sizes: %s%s \r", Cyan(completed), incomplete)
	it := fmt.Sprintf("%s ", v.Aspect)
	completed += it
	incomplete = strings.Replace(incomplete, it, "", 1)
	cmd := v.mkCmd(cmdtemplate)
	chkExec(cmd)
	v.readRate()
	fmt.Printf(" %s variant sizes: %s%s \r", Cyan("."), Cyan(completed), incomplete)
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	stanza := fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v,CODECS=\"%s\"", v.Bandwidth, v.Aspect, mastercodec)
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
	one := "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"webvtt\","
	two := "NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,"
	line :=mkLine()
	three := fmt.Sprintf("LANGUAGE=\"en\",URI=\"%ssubs/vtt_index.m3u8\"\n",line)
	return one + two + three
}

func mkIncomplete(variants []Variant) {
	for _, v := range variants {
		incomplete += fmt.Sprintf("%s ", v.Aspect)
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
func mkTopLevel() {
	if toplevel == "" {
		toplevel = strings.Split(infile, `.`)[0]
		if strings.Contains(toplevel,"/"){
		one:=strings.Split(toplevel,"/")
		toplevel=one[len(one)-1]
	}
	}
	os.MkdirAll(toplevel, 0755)
}

//Extract 608 captions to an webvtt file.
func extractCaptions() string {
	fmt.Printf("%s caption file : %s \n", Cyan(" ."), Cyan(infile))
	fmt.Printf(" . %s", Cyan("extracting captions \r"))
	subfile := fmt.Sprintf("%s/%s.vtt",toplevel, toplevel)
	prefix := "ffmpeg -y -f lavfi -fix_sub_duration "
	postfix := fmt.Sprintf("-i movie=%s[out0+subcc] %s", infile, subfile)
	cmd := prefix + postfix
	chkExec(cmd)
	return subfile
}

// Extract captions to segment,
// unless a subtitle file is passed in with "-s"
func mkSubfile() {
	addsubs = false
	if !(webvtt) {
		if (subfile == "") && (hasCaptions()) {
			subfile = extractCaptions()
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
	fixUrlPrefix()
	fmt.Println(Cyan(" ."), "video file   :", Cyan(infile))
	fmt.Println(Cyan(" ."), "toplevel dir :", Cyan(toplevel))
	mkSubfile()
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", toplevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	fmt.Println(Cyan(" ."), "subtitle file:", Cyan(subfile))
	mkIncomplete(variants)
	for _, v := range variants {
		v.start()
		if addsubs && !(webvtt) {
			mvCaptions(v.Name)
			w.WriteString(mkSubStanza())
			webvtt = true
		}
		w.WriteString(fmt.Sprintf("%s\n", v.mkStanza()))
		line :=mkLine()
		w.WriteString(fmt.Sprintf("%s%s/index.m3u8\n",line, v.Name))
	}
	w.Flush()
	fmt.Println()
}

func stamp() {
	t := time.Now()
	fmt.Println(Cyan(" ."), Cyan(t.Format(time.Stamp)))
}

func runBatch() {
	batch = strings.Replace(batch, " ", ",", -1)
	splitbatch := strings.Split(batch, ",")
	for i, b := range splitbatch {
		fmt.Println("\n", Cyan(i+1), "of", len(splitbatch))
		stamp()
		webvtt = false
		subfile = ""
		infile = b
		completed = ""
		toplevel = ""
		mkTopLevel()
		variants := dataToVariants()
		mkAll(variants)
	}
}
func mkLine() string {
	line :=urlprefix
		if urlprefix !="" {
			line+=fmt.Sprintf("%s/",toplevel)
		}
	return line
}	

func fixUrlPrefix(){
	
	if (urlprefix !="") && !(strings.HasSuffix(urlprefix,"/")) {
		urlprefix +="/"
	}
}

func do() {
	mkFlags()
	fixUrlPrefix()
	if batch != "" {
		runBatch()
	} else {
		if infile != "" {
			stamp()
			variants := dataToVariants()
			mkAll(variants)
		} else {
			flag.PrintDefaults()
		}
	}

}

func mkFlags() {
	flag.StringVar(&infile, "i", "", "Video file to segment (either -i or -b is required)")
	flag.StringVar(&subfile, "s", "", "subtitle file to segment (optional)")
	flag.StringVar(&toplevel, "d", "", "override top level directory for hls files (optional)")
	flag.StringVar(&jasonfile, "j", `./hls.json`, "JSON file of variants (optional)")
	flag.StringVar(&cmdtemplate, "t", `./cmd.template`, "command template file (optional)")
	flag.StringVar(&batch, "b", "", "batch mode, list multiple input files (either -i or -b is required)")
	flag.StringVar(&urlprefix,"u","","url prefix to add to index.m3u8 path in master.m3u8 (optional)")
	flag.Parse()
}

func main() {
	do()

}
