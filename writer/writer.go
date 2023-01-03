package writer

import (
	"encoding/binary"
	"errors"
	"io"
	"log"
	"net"
	"pion-webrtc-sfu/configuration"
	"pion-webrtc-sfu/sessions"
)

// write incoming video UDP packets to video track
func StartVideoWriterLoop(config *configuration.Configuration) {
	inboundRTPPacket := make([]byte, config.Rtc_receive_rtp_buffsize) // UDP MTU
	// open UDP listener
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: int(config.Rtc_video_tracks_receive_port)})
	if err != nil {
		log.Fatalf("could not open UDP port for video listener: (%v)\n", err)
	}
	// read from listener and write to track if ssrc matches an existing session
	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			log.Fatalf("error trying to read from video UDP listener: %v\n", err)
		}
		// extract ssrc from bytestream
		stream_ssrc := binary.BigEndian.Uint32(inboundRTPPacket[:n][8:12])
		// write to session track if exists
		sess := sessions.ReturnSessionByIdIfExists(stream_ssrc)
		if sess != nil {
			if _, err = sess.TrackGroup.VideoTrack.Write(inboundRTPPacket[:n]); err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					continue
				}
				log.Printf("video track of session %v writer got exception: %s\n", stream_ssrc, err)
			}
		}
	}
}

// write incoming audio UDP packets to audio track
func StartAudioWriterLoop(config *configuration.Configuration) {
	inboundRTPPacket := make([]byte, config.Rtc_receive_rtp_buffsize) // UDP MTU
	// open UDP listener
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("0.0.0.0"), Port: int(config.Rtc_audio_tracks_receive_port)})
	if err != nil {
		log.Fatalf("could not open UDP port for audio listener: (%v)\n", err)
	}
	// read from listener and write to track if ssrc matches an existing session
	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			log.Fatalf("error trying to read from audio UDP listener: %v\n", err)
		}
		// extract ssrc from bytestream
		stream_ssrc := binary.BigEndian.Uint32(inboundRTPPacket[:n][8:12])
		// write to session track if exists
		sess := sessions.ReturnSessionByIdIfExists(stream_ssrc)
		if sess != nil {
			if _, err = sess.TrackGroup.AudioTrack.Write(inboundRTPPacket[:n]); err != nil {
				if errors.Is(err, io.ErrClosedPipe) {
					continue
				}
				log.Printf("audio track of session %v writer got exception: %s\n", stream_ssrc, err)
			}
		}
	}
}
