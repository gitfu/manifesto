package main

import (
	"encoding/json"
	"fmt"
	
)

func main() {
	
jason :=[]byte (`
 {
    "streams": [
        {
            "index": 0,
            "codec_name": "h264",
            "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
            "profile": "Main",
            "codec_type": "video",
            "codec_time_base": "1001/60000",
            "codec_tag_string": "[27][0][0][0]",
            "codec_tag": "0x001b",
            "width": 1280,
            "height": 720,
            "coded_width": 1280,
            "coded_height": 720,
            "has_b_frames": 1,
            "sample_aspect_ratio": "1:1",
            "display_aspect_ratio": "16:9",
            "pix_fmt": "yuv420p",
            "level": 31,
            "chroma_location": "left",
            "field_order": "progressive",
            "refs": 1,
            "is_avc": "false",
            "nal_length_size": "0",
            "id": "0x1e1",
            "r_frame_rate": "30000/1001",
            "avg_frame_rate": "30000/1001",
            "time_base": "1/90000",
            "start_pts": 183003,
            "start_time": "2.033367",
            "bits_per_raw_sample": "8",
            "disposition": {
                "default": 0,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0
            }
        },
        {
            "index": 1,
            "codec_name": "aac",
            "codec_long_name": "AAC (Advanced Audio Coding)",
            "profile": "LC",
            "codec_type": "audio",
            "codec_time_base": "1/48000",
            "codec_tag_string": "[15][0][0][0]",
            "codec_tag": "0x000f",
            "sample_fmt": "fltp",
            "sample_rate": "48000",
            "channels": 2,
            "channel_layout": "stereo",
            "bits_per_sample": 0,
            "id": "0x1e2",
            "r_frame_rate": "0/0",
            "avg_frame_rate": "0/0",
            "time_base": "1/90000",
            "start_pts": 181083,
            "start_time": "2.012033",
            "duration_ts": 154512000,
            "duration": "1716.800000",
            "bit_rate": "256875",
            "disposition": {
                "default": 0,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0
            },
            "tags": {
                "language": "eng"
            }
        },
        {
            "index": 2,
            "codec_name": "scte_35",
            "codec_long_name": "SCTE 35 Message Queue",
            "codec_type": "data",
            "codec_tag_string": "[0][0][0][0]",
            "codec_tag": "0x0000",
            "id": "0x1f4",
            "r_frame_rate": "0/0",
            "avg_frame_rate": "0/0",
            "time_base": "1/90000",
            "start_pts": 181083,
            "start_time": "2.012033",
            "duration_ts": 154512000,
            "duration": "1716.800000",
            "disposition": {
                "default": 0,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0
            }
        },
        {
            "index": 3,
            "codec_name": "timed_id3",
            "codec_long_name": "timed ID3 metadata",
            "codec_type": "data",
            "codec_tag_string": "ID3 ",
            "codec_tag": "0x20334449",
            "id": "0x1f6",
            "r_frame_rate": "0/0",
            "avg_frame_rate": "0/0",
            "time_base": "1/90000",
            "start_pts": 183003,
            "start_time": "2.033367",
            "disposition": {
                "default": 0,
                "dub": 0,
                "original": 0,
                "comment": 0,
                "lyrics": 0,
                "karaoke": 0,
                "forced": 0,
                "hearing_impaired": 0,
                "visual_impaired": 0,
                "clean_effects": 0,
                "attached_pic": 0,
                "timed_thumbnails": 0
            }
        }
    ],
    "format": {
        "filename": "/home/a/bitstream/examples/scte35.ts",
        "nb_streams": 4,
        "nb_programs": 1,
        "format_name": "mpegts",
        "format_long_name": "MPEG-TS (MPEG-2 Transport Stream)",
        "start_time": "2.012033",
        "duration": "1716.800000",
        "size": "940732448",
        "bit_rate": "4383655",
        "probe_score": 50
    }
}
`)

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
Acodec		string
}

var st Stanza						
var f Container
json.Unmarshal(jason, &f)

st.Bandwidth=f.Format.BitRate
for _,i := range f.Streams{
	
	if i.CodecType=="video" {
		st.Resolution= fmt.Sprintf("=%vx%v",i.Width,i.Height)
		if i.Profile=="High"{ 
			st.Profile="64"
		}
		if i.Profile =="Main"{
			st.Profile="4d"
		}		
		if i.Profile =="Baseline"{
			st.Profile="42"
		}		
		st.Level=i.Level
	
		}
	if i.CodecType=="audio" {
		if i.Profile =="HE-AACv2"{
			st.Acodec="mp4a.40.5"
		}
		if i.Profile =="LC" {
			st.Acodec="mp4a.40.2"
		}
	}
}
fmt.Printf("#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%v,RESOLUTION=%s,CODECS=\"avc1.%v00%x,%v\"\n",st.Bandwidth,st.Resolution,st.Profile,int(st.Level),st.Acodec)

}
