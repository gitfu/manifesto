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

var x264level = "3.0"
var x264profile = "high"
var mastercodec = "avc1.64001E,mp4a.40.2"

type Job struct {
	InFile      string
	SubFile     string
	AddSubs     bool
	TopLevel    string
	WebVtt      bool
	JasonFile   string
	CmdTemplate string
	Incomplete  string
	Completed   string
	UrlPrefix   string
	Variants    []Variant
}

func (j *Job) mkFlags() {
	flag.StringVar(&j.InFile, "i", "", "Video file to segment (either -i or -b is required)")
	flag.StringVar(&j.SubFile, "s", "", "subtitle file to segment (optional)")
	flag.StringVar(&j.TopLevel, "d", "", "override top level directory for hls files (optional)")
	flag.StringVar(&j.JasonFile, "j", `./hls.json`, "JSON file of variants (optional)")
	flag.StringVar(&j.CmdTemplate, "t", `./cmd.template`, "command template file (optional)")
	flag.StringVar(&j.UrlPrefix, "u", "", "url prefix to add to index.m3u8 path in master.m3u8 (optional)")
	flag.Parse()
}

// Read json file for variants
func (j *Job) dataToVariants() {
	data, err := ioutil.ReadFile(j.JasonFile)
	chk(err, "Error reading JSON file")
	json.Unmarshal(data, &j.Variants)
}

// Set the TopLeveldir for variants by splitting video file name at the "."
func (j *Job) mkTopLevel() {
	if j.TopLevel == "" {
		j.TopLevel = strings.Split(j.InFile, `.`)[0]
		if strings.Contains(j.TopLevel, "/") {
			one := strings.Split(j.TopLevel, "/")
			j.TopLevel = one[len(one)-1]
		}
	}
	os.MkdirAll(j.TopLevel, 0755)
}

func (j *Job) mkIncomplete() {
	for _, v := range j.Variants {
		j.Incomplete += fmt.Sprintf("%s ", v.Aspect)
	}
}

//Extract 608 captions to an WebVtt file.
func (j *Job) extractCaptions() {
	fmt.Printf("%s caption file : %s \n", Cyan(" ."), Cyan(j.InFile))
	fmt.Printf(" . %s", Cyan("extracting captions \r"))
	j.SubFile = fmt.Sprintf("%s/%s.vtt", j.TopLevel, j.TopLevel)
	prefix := "ffmpeg -y -f lavfi -fix_sub_duration "
	postfix := fmt.Sprintf("-i movie=%s[out0+subcc] %s", j.InFile, j.SubFile)
	cmd := prefix + postfix
	chkExec(cmd)

}

// probes for Closed Captions in video file.
func (j *Job) hasCaptions() bool {
	cmd := fmt.Sprintf("ffprobe -i %s", j.InFile)
	data := chkExec(cmd)
	if strings.Contains(data, "Captions") {
		return true
	}
	return false
}

// Captions are segmented along with the first variant and then moved to toplevel/subs
func (j *Job) mvSubtitles(vardir string) {
	srcdir := fmt.Sprintf("%s/%s", j.TopLevel, vardir)
	destdir := fmt.Sprintf("%s/subs", j.TopLevel)
	os.MkdirAll(destdir, 0755)
	files, err := ioutil.ReadDir(srcdir)
	chk(err, "Error moving Captions")
	for _, f := range files {
		if strings.Contains(f.Name(), "vtt") {
			os.Rename(fmt.Sprintf("%s/%s", srcdir, f.Name()), fmt.Sprintf("%s/%s", destdir, f.Name()))
		}
	}
}

// Extract captions to segment,
// unless a subtitle file is passed in with "-s"
func (j *Job) mkSubfile() {
	j.AddSubs = false
	if !(j.WebVtt) {
		if (j.SubFile == "") && (j.hasCaptions()) {
			j.extractCaptions()
		}
		if j.SubFile != "" {
			j.AddSubs = true
		}
	}
}

// create a subtitle stanza for use in the  master.m3u8
func (j *Job) mkSubStanza() string {
	one := "#EXT-X-MEDIA:TYPE=SUBTITLES,GROUP-ID=\"WebVtt\","
	two := "NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,FORCED=NO,"
	line := j.mkLine()
	three := fmt.Sprintf("LANGUAGE=\"en\",URI=\"%ssubs/vtt_index.m3u8\"\n", line)
	return one + two + three
}

func (j *Job) doVariant(v Variant){
		v.job = j
		v.start()
}

// Make all variants and write master.m3u8
func (j *Job) mkAll() {
	fmt.Println(Cyan(" ."), "video file   :", Cyan(j.InFile))
	fmt.Println(Cyan(" ."), "TopLeveldir :", Cyan(j.TopLevel))
	j.mkSubfile()
	var m3u8Master = fmt.Sprintf("%s/master.m3u8", j.TopLevel)
	fp, err := os.Create(m3u8Master)
	chk(err, "in mkAll")
	defer fp.Close()
	w := bufio.NewWriter(fp)
	w.WriteString("#EXTM3U\n")
	fmt.Println(Cyan(" ."), "subtitle file:", Cyan(j.SubFile))
	j.mkIncomplete()
	for _, v := range j.Variants {
		j.doVariant(v)
		if j.AddSubs && !(j.WebVtt) {
			j.mvSubtitles(v.Name)
			w.WriteString(j.mkSubStanza())
			j.WebVtt = true
		}
		w.WriteString(fmt.Sprintf("%s\n", v.mkStanza()))
		line := j.mkLine()
		w.WriteString(fmt.Sprintf("%s%s/index.m3u8\n", line, v.Name))
		w.Flush()
	}
	fmt.Println()
}

func (j *Job) mkLine() string {
	line := j.UrlPrefix
	if j.UrlPrefix != "" {
		line += fmt.Sprintf("%s/", j.TopLevel)
	}
	return line
}

func (j *Job) fixUrlPrefix() {
	if (j.UrlPrefix != "") && !(strings.HasSuffix(j.UrlPrefix, "/")) {
		j.UrlPrefix += "/"
	}
}

func (j *Job) do() {
	j.mkFlags()
	if j.InFile != "" {
		j.mkTopLevel()
		j.fixUrlPrefix()
		j.dataToVariants()
		j.mkAll()
	} else {
		flag.PrintDefaults()
	}
}

// End Job

// Variant struct for HLS variants
type Variant struct {
	job       *Job
	Name      string `json:"name"`
	Aspect    string `json:"aspect"`
	Rate      string `json:"framerate"`
	Vbr       string `json:"vbitrate"`
	Abr       string `json:"abitrate"`
	Buf       string `json:"bufsize"`
	Bandwidth string
}

// Create variant's destination directory
func (v *Variant) mkDest() string {
	dest := fmt.Sprintf("%s/%s", v.job.TopLevel, v.Name)
	os.MkdirAll(dest, 0755)
	return dest
}

func (v *Variant) mkInputs() string {
	inputs := fmt.Sprintf(" -i %s", v.job.InFile)
	if v.job.AddSubs && !(v.job.WebVtt) {
		inputs = fmt.Sprintf(" -i %s -i %s  ", v.job.InFile, v.job.SubFile)
	}
	return inputs
}

// This Variant method assembles the ffmpeg command
func (v *Variant) mkCmd(CmdTemplate string) string {
	data, err := ioutil.ReadFile(CmdTemplate)
	chk(err, "Error reading template file")
	inputs := v.mkInputs()
	r := strings.NewReplacer("INPUTS", inputs, "ASPECT", v.Aspect,
		"VBITRATE", v.Vbr,"BUFSIZE",v.Buf, "X264LEVEL", x264level,
		"X264PROFILE", x264profile, "FRAMERATE", v.Rate,
		"ABITRATE", v.Abr, "TOPLEVEL", v.job.TopLevel,
		"NAME", v.Name, "\n", " ")
	cmd := fmt.Sprintf("%s\n", r.Replace(string(data)))
	return cmd
}

// Read actual bitrate from first segment to set bandwidth in master.m3u8
func (v *Variant) readRate() {
	cmd := fmt.Sprintf("ffprobe -i %s/%s/index0.ts", v.job.TopLevel, v.Name)
	data := chkExec(cmd)
	two := strings.Split(data, "bitrate: ")[1]
	rate := strings.Split(two, " kb/s")[0]
	v.Bandwidth = fmt.Sprintf("%v000", rate)
}

// Start transcoding the variant
func (v *Variant) start() {
	v.mkDest()
	fmt.Printf(" . variant sizes: %s%s \r", Cyan(v.job.Completed), v.job.Incomplete)
	it := fmt.Sprintf("%s ", v.Aspect)
	v.job.Completed += it
	v.job.Incomplete = strings.Replace(v.job.Incomplete, it, "", 1)
	cmd := v.mkCmd(v.job.CmdTemplate)
	chkExec(cmd)
	v.readRate()
	fmt.Printf(" %s variant sizes: %s%s \r", Cyan("."), Cyan(v.job.Completed), v.job.Incomplete)
}

// #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=7483000,RESOLUTION=1920:1080,
// hd1920/index.m3u8
func (v *Variant) mkStanza() string {
	stanza := fmt.Sprintf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%v,CODECS=\"%s\"", v.Bandwidth, v.Aspect, mastercodec)
	if v.job.AddSubs {
		stanza = fmt.Sprintf("%s,SUBTITLES=\"WebVtt\"", stanza)
	}
	return stanza
}

// End Variant

func chkExec(cmd string) string {
	// Executes external commands and checks for runtime errors
	parts := strings.Fields(cmd)
	data, err := exec.Command(parts[0], parts[1:]...).CombinedOutput()
	chk(err, fmt.Sprintf("Error running \n %s \n %v", cmd, string(data)))
	return string(data)
}

// Generic catchall error checking
func chk(err error, mesg string) {
	if err != nil {
		fmt.Printf("%s\n", mesg)
		//panic(err)
	}
}

func stamp() {
	t := time.Now()
	fmt.Println(Cyan(" ."), Cyan(t.Format(time.Stamp)))
}

func main() {
	stamp()
	var j Job
	j.do()

}
