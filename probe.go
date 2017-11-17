package main

import (
	"encoding/json"
	"fmt"
	
)


x264Profiles := map[string]string{"Baseline":"42","Main": "4d","High": "64"}
AudioProfiles := map[string]string{"HE-AACv2":"mp4a.40.5","LC":"mp4a.40.2"}
	

type Format struct{
	FormatName 	string	`json:"format_name"`
	Duration	string	`json:"duration"`
	BitRate		string	`json:"bit_rate"`
}	

type Stream struct {
CodecType 	string 	`json:"codec_type"`
CodecName	string 	`json:"codec_name"`
Profile 	string	`json:"profile"`	
Level		float64	`json:"level"`
Width		float64	`json:"width"`
Height		float64	`json:"height"`	
	
}		

type Container struct {
Streams	[]Stream	`json:"streams"`
Format	Format		`json:"format"`	
}	

type Stanza struct { 
Bandwidth	string
Resolution	string
Level		float64
Profile		string
AProfile	string
}

	
var st Stanza						
var f Container
json.Unmarshal(jason, &f)

st.Bandwidth=f.Format.BitRate
for _,i := range f.Streams{	
	if i.CodecType=="video" {
		st.Resolution= fmt.Sprintf("=%vx%v",i.Width,i.Height)
		st.Profile=x264Profiles[i.Profile]
		st.Level=i.Level
		}
	if i.CodecType=="audio" {
		st.AProfile=","+AudioProfiles[i.Profile]	
	}
}
fmt.Printf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%s,CODECS=\"avc1.%v00%x%v\"\n",st.Bandwidth,st.Resolution,st.Profile,int(st.Level),st.AProfile)



