package tracks

import (
	"fmt"
	"math/rand"
	"pion-webrtc-sfu/configuration"
	"time"

	"github.com/pion/randutil"
	"github.com/pion/webrtc/v3"
)

/*
a track group contains all the information that should be
known about incoming stream group (video/audio)
*/
type TrackGroup struct {
	VideoTrack *webrtc.TrackLocalStaticRTP
	AudioTrack *webrtc.TrackLocalStaticRTP
}

/*
create new track group for video/audio streams
extracts audio/video codec types from configuration
*/
func NewTrackGroup(conf *configuration.Configuration) (TrackGroup, error) {
	rand.Seed(time.Now().UTC().UnixNano())
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: conf.Rtc_video_codec},
		fmt.Sprintf("video-%d", randutil.NewMathRandomGenerator().Uint32()),
		fmt.Sprintf("video-%d", randutil.NewMathRandomGenerator().Uint32()),
	)
	if err != nil {
		return TrackGroup{}, err
	}
	audioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: conf.Rtc_audio_codec},
		fmt.Sprintf("audio-%d", randutil.NewMathRandomGenerator().Uint32()),
		fmt.Sprintf("audio-%d", randutil.NewMathRandomGenerator().Uint32()),
	)
	if err != nil {
		return TrackGroup{}, err
	}
	return TrackGroup{
		VideoTrack: videoTrack,
		AudioTrack: audioTrack,
	}, nil
}
