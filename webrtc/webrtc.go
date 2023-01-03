package webrtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"pion-webrtc-sfu/configuration"
	"pion-webrtc-sfu/sessions"
	"pion-webrtc-sfu/user"
	"strings"

	"github.com/pion/interceptor"
	"github.com/pion/rtcp"
	pwrtc "github.com/pion/webrtc/v3"
)

type WebrtcClient struct {
	usr            *user.User               // contains all required info
	peerConnection *pwrtc.PeerConnection    // peerconnection instance
	videoRTPSender *pwrtc.RTPSender         // audio rtp sender for rtcp parsing
	audioRTPSender *pwrtc.RTPSender         // vide rtp sender for rtcp parsing
	currentOffer   pwrtc.SessionDescription // current offer
}

// creates webrtc object for server-client communication
func NewWebrtcClient(config *configuration.Configuration, usr *user.User, sessionID uint32) (*WebrtcClient, error) {
	wrtcclient := WebrtcClient{
		usr: usr,
	}
	// create peerconnection
	err := wrtcclient.createPeerConnection(config, sessionID, usr.Uuid)
	if err != nil {
		// close peerconnection if it was created before error
		if wrtcclient.peerConnection != nil {
			if err := wrtcclient.peerConnection.Close(); err != nil {
				log.Printf("user %v in session %v could not close pc: %v\n", usr.Uuid, sessionID, err)
			}
		}
		wrtcclient.usr.State.SetRtcState(user.Done)
		return nil, err
	}
	return &wrtcclient, nil
}

// required for processing things like NACK (video)
func (wc *WebrtcClient) processVRTCP() {
	rtcpBuf := make([]byte, 1500)
	for wc.usr.State.GetRtcState() != user.Done {
		n, _, rtcpErr := wc.videoRTPSender.Read(rtcpBuf)
		if rtcpErr != nil {
			break
		}
		t, err := rtcp.Unmarshal(rtcpBuf[:n])
		if err != nil {
			continue
		}
		for _, p := range t {
			/*
				some other RTCP types need to be managed without Pion
				PIL for example needs to be communicated to the encoder directyly
			*/
			switch p := p.(type) {
			case *rtcp.PictureLossIndication:
			case *rtcp.FullIntraRequest:
			case *rtcp.ReceiverEstimatedMaximumBitrate:
			case *rtcp.ReceiverReport:
			case *rtcp.SenderReport:
			case *rtcp.SliceLossIndication:
			case *rtcp.TransportLayerNack:
			default:
				var pbyte []byte
				p.Unmarshal(pbyte)
				log.Printf("Could not parse rtp packet, skipping: (%v)\n", string(pbyte))
			}
		}
	}
}

// required for processing things like NACK (audio)
func (wc *WebrtcClient) processARTCP() {
	rtcpBuf := make([]byte, 1500)
	for wc.usr.State.GetRtcState() != user.Done {
		n, _, rtcpErr := wc.audioRTPSender.Read(rtcpBuf)
		if rtcpErr != nil {
			break
		}
		t, err := rtcp.Unmarshal(rtcpBuf[:n])
		if err != nil {
			continue
		}
		for _, p := range t {
			/*
				some other RTCP types need to be managed without Pion
				PIL for example needs to be communicated to the encoder directyly
			*/
			switch p := p.(type) {
			case *rtcp.PictureLossIndication:
			case *rtcp.FullIntraRequest:
			case *rtcp.ReceiverEstimatedMaximumBitrate:
			case *rtcp.ReceiverReport:
			case *rtcp.SenderReport:
			case *rtcp.SliceLossIndication:
			case *rtcp.TransportLayerNack:
			default:
				var pbyte []byte
				p.Unmarshal(pbyte)
				log.Printf("Could not parse rtp packet, skipping: (%v)\n", string(pbyte))
			}
		}
	}
}

func (wc *WebrtcClient) createPeerConnection(config *configuration.Configuration, sessionID uint32, userID string) error {
	var err error
	//		create pion API		//
	settingsEngine := &pwrtc.SettingEngine{}
	// set valid port range
	settingsEngine.SetEphemeralUDPPortRange(config.Server_ephemeral_udp_port_range.Start, config.Server_ephemeral_udp_port_range.End)
	nips := strings.Split(config.Server_NAT_1to1_IPs, ",")
	if nips[0] != "" {
		settingsEngine.SetNAT1To1IPs(
			nips,
			pwrtc.ICECandidateTypeHost,
		)
	}
	// assuming you have a static IP. simplifies things
	settingsEngine.SetLite(true)
	settingsEngine.SetICETimeouts(wc.usr.Settings.RTCDisconnectTimeout, wc.usr.Settings.RTCFailedTimeout, wc.usr.Settings.RTCKeepaliveInterval)
	mediaEngine := &pwrtc.MediaEngine{}
	// register all available codecs -- might be neater to only register required ones
	// take a look at the called function to learn exactly how
	if err = mediaEngine.RegisterDefaultCodecs(); err != nil {
		return err
	}
	// register default(all) interceptors - generating reports and NACK handling currently
	interceptorRegistry := &interceptor.Registry{}
	if err = pwrtc.RegisterDefaultInterceptors(mediaEngine, interceptorRegistry); err != nil {
		return err
	}
	// create api
	api := pwrtc.NewAPI(pwrtc.WithMediaEngine(mediaEngine), pwrtc.WithSettingEngine(*settingsEngine), pwrtc.WithInterceptorRegistry(interceptorRegistry))
	configuration := pwrtc.Configuration{
		BundlePolicy:  pwrtc.BundlePolicyBalanced,
		RTCPMuxPolicy: pwrtc.RTCPMuxPolicyRequire,
	}
	if wc.peerConnection, err = api.NewPeerConnection(configuration); err != nil {
		return err
	}
	if wc.peerConnection.ConnectionState() != pwrtc.PeerConnectionStateNew {
		return fmt.Errorf("createPeerConnection called in non-new state (%s), exiting", wc.peerConnection.ConnectionState())
	}
	//		set event handlers		//
	wc.peerConnection.OnNegotiationNeeded(func() {
		// logging if needed
	})
	wc.peerConnection.OnICEConnectionStateChange(func(is pwrtc.ICEConnectionState) {
		// logging if needed
	})
	wc.peerConnection.OnICEGatheringStateChange(func(is pwrtc.ICEGathererState) {
		// logging if needed
	})
	wc.peerConnection.OnSignalingStateChange(func(ss pwrtc.SignalingState) {
		// logging if needed
	})
	// send generated ICE candidates to client buffer - which is later sent to the client over websocket
	wc.peerConnection.OnICECandidate(func(i *pwrtc.ICECandidate) {
		if i == nil {
			return
		}
		iceCandidate, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Printf("user %v in session %v could not marshal icecandidate payload\n", wc.usr.Uuid, sessionID)
			return
		}
		// notify client
		wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
			Type:       user.MESSAGE_ICECANDIDATE,
			RawPayload: iceCandidate,
		})
	})
	//		track config		//
	// get tracks
	sess := sessions.ReturnSessionByIdIfExists(sessionID)
	if sess == nil {
		return errors.New("tracks do not exist in sessions")
	}
	// add tracks and set RTPSenders
	wc.videoRTPSender, err = wc.peerConnection.AddTrack(sess.TrackGroup.VideoTrack)
	if err != nil {
		return err
	}
	wc.audioRTPSender, err = wc.peerConnection.AddTrack(sess.TrackGroup.AudioTrack)
	if err != nil {
		return err
	}
	//		create offer		//
	offerOptions := pwrtc.OfferOptions{
		OfferAnswerOptions: pwrtc.OfferAnswerOptions{
			VoiceActivityDetection: false,
		},
		// initial offer doesn't have this
		ICERestart: false,
	}
	offer, err := wc.peerConnection.CreateOffer(&offerOptions)
	if err != nil {
		return err
	}
	wc.currentOffer = offer
	/*
		set the callback handler for peer connection state
		+ notify websocket when the peer has connected/disconnected
	*/
	wc.peerConnection.OnConnectionStateChange(func(s pwrtc.PeerConnectionState) {
		log.Printf("user %v in session %v pc state has changed: %v\n", userID, sessionID, s.String())
		if s == pwrtc.PeerConnectionStateDisconnected {
			if wc.usr.State.GetRtcState() != user.Connected && wc.peerConnection.ConnectionState() != pwrtc.PeerConnectionStateClosed {
				// not connected, dont need ice restart
				return
			}
			// reset state and send new offer
			wc.usr.State.SetRtcState(user.AwaitingConnection)
			log.Printf("user %v in session %v pc attempting restart\n", userID, sessionID)
			// create new sdp
			offerOptions := pwrtc.OfferOptions{
				OfferAnswerOptions: pwrtc.OfferAnswerOptions{
					VoiceActivityDetection: false,
				},
				ICERestart: true,
			}
			offer, err := wc.peerConnection.CreateOffer(&offerOptions)
			if err != nil {
				log.Printf("user %v in session %v couldn't create offer for restart: %v\n", userID, sessionID, err)
				return
			}
			wc.currentOffer = offer
			// notify server
			wc.usr.RtcMessageBuffer.PushToServerBuffer(user.Message{Type: user.MESSAGE_ICERESTART})
		} else if s == pwrtc.PeerConnectionStateFailed {
			log.Printf("user %v in session %v failed\n", userID, sessionID)
			// notify server
			wc.usr.RtcMessageBuffer.PushToServerBuffer(user.Message{Type: user.MESSAGE_PCFAILED})
			wc.usr.State.KillRtc()
		} else if s == pwrtc.PeerConnectionStateConnected {
			wc.usr.State.SetRtcState(user.Connected)
		}
	})
	wc.usr.State.SetRtcState(user.AwaitingConnection)
	return nil
}

// close webrtc client and update user sessions
func (wc *WebrtcClient) Close() error {
	err := wc.peerConnection.Close()
	wc.usr.State.SetRtcState(user.Done)
	// update session and close if needed
	sessions.UpdateSessions()
	return err
}

// main loop
func (wc *WebrtcClient) Loop(config *configuration.Configuration, sessionID uint32) {
	defer wc.Close()
	// start rtcp readers
	go wc.processVRTCP()
	go wc.processARTCP()
	// blocking loop
	for {
		select {
		case <-wc.usr.State.WsKilled():
			log.Printf("user %v in session %v killed ws, also killing rtc\n", wc.usr.Uuid, sessionID)
			return
		case <-wc.usr.State.RtcKilled():
			log.Printf("user %v in session %v killed rtc\n", wc.usr.Uuid, sessionID)
			return
		// requests made to the server by the server or client
		case sockMsg := <-wc.usr.RtcMessageBuffer.ReadFromServerBuffer():
			switch sockMsg.Type {
			// start webrtc - works by setting the local description and kickstarting the process
			case user.MESSAGE_STARTRTC:
				if wc.peerConnection.SignalingState() == pwrtc.SignalingStateHaveLocalOffer {
					log.Printf("user %v in session %v ignoring startrtc because offer is already set\n", wc.usr.Uuid, sessionID)
					continue
				}
				if wc.peerConnection.ConnectionState() == pwrtc.PeerConnectionStateConnected {
					log.Printf("user %v in session %v ignoring startrtc webrtc already connected\n", wc.usr.Uuid, sessionID)
					continue
				}
				if err := wc.peerConnection.SetLocalDescription(wc.currentOffer); err != nil {
					// notify error to client
					wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
						Type:       user.MESSAGE_PCFAILED,
						RawPayload: []byte(fmt.Sprintf("%q", err.Error())),
					})
					continue
				}
				// marshal offer
				offerjson, err := json.Marshal(*wc.peerConnection.LocalDescription())
				if err != nil {
					// notify error to client
					wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
						Type:       user.MESSAGE_PCFAILED,
						RawPayload: []byte(fmt.Sprintf("%q", err.Error())),
					})
					return
				}
				// tell client the offer
				wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
					Type:       user.MESSAGE_SDP,
					RawPayload: offerjson,
				})
			// server received sdp (from remote client)
			case user.MESSAGE_SDP:
				var sdp pwrtc.SessionDescription
				if err := json.Unmarshal(sockMsg.RawPayload, &sdp); err != nil {
					log.Printf("user %v in session %v could not unmarshal sdp payload\n", wc.usr.Uuid, sessionID)
					continue
				}
				if err := wc.peerConnection.SetRemoteDescription(sdp); err != nil {
					// notify error to client
					wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
						Type:       user.MESSAGE_PCFAILED,
						RawPayload: []byte(fmt.Sprintf("%q", err.Error())),
					})
					continue
				}
			// server received ice candidate (from remote client)
			case user.MESSAGE_ICECANDIDATE:
				var icecandidate pwrtc.ICECandidate
				if err := json.Unmarshal(sockMsg.RawPayload, &icecandidate); err != nil {
					log.Printf("user %v in session %v could not unmarshal icecandidate payload\n", wc.usr.Uuid, sessionID)
					continue
				}
				wc.peerConnection.AddICECandidate(icecandidate.ToJSON())
			// server received ice restart request (from connectionStateChange callback)
			case user.MESSAGE_ICERESTART:
				if err := wc.peerConnection.SetLocalDescription(wc.currentOffer); err != nil {
					// notify error to client
					wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
						Type:       user.MESSAGE_PCFAILED,
						RawPayload: []byte(fmt.Sprintf("%q", err.Error())),
					})
					continue
				}
				offerjson, err := json.Marshal(*wc.peerConnection.LocalDescription())
				if err != nil {
					log.Printf("user %v in session %v could not marshal sdp payload\n", wc.usr.Uuid, sessionID)
					continue
				}
				// tell client the new offer
				wc.usr.WsMessageBuffer.PushToClientBuffer(user.Message{
					Type:       user.MESSAGE_SDP,
					RawPayload: offerjson,
				})
			default:
				log.Printf("user %v in session %v got bad payload type in server\n", wc.usr.Uuid, sessionID)
			}
		// requests to make to the client
		case sockMsg := <-wc.usr.RtcMessageBuffer.ReadFromClientBuffer():
			switch sockMsg.Type {
			// no datachannels opened at the moment.
			// if datachannels are used client messages can be parsed and handled from here
			default:
				log.Printf("user %v in session %v got bad payload type in client\n", wc.usr.Uuid, sessionID)
			}
		}
	}
}
