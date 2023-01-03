package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"pion-webrtc-sfu/configuration"
	"pion-webrtc-sfu/sessions"
	"pion-webrtc-sfu/user"
	"pion-webrtc-sfu/webrtc"
	"time"

	"github.com/gin-gonic/gin"

	gorillaSocket "github.com/gorilla/websocket"
	pwrtc "github.com/pion/webrtc/v3"
)

type WebsocketClient struct {
	usr  *user.User          // connected user
	sock *gorillaSocket.Conn // websocket object used by client

}

func NewWebsocketClient(c *gin.Context, usr *user.User) (*WebsocketClient, error) {
	// create websocket client
	// TODO: origin check disabled, might need to enable in other situations
	upgrader := gorillaSocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	sock, err := upgrader.Upgrade((*c).Writer, (*c).Request, nil)
	if err != nil {
		usr.State.SetWsState(user.Done)
		return nil, err
	}
	usr.State.SetWsState(user.Connected)
	return &WebsocketClient{
		usr:  usr,
		sock: sock,
	}, nil
}

// send payload to websocket client
// return errors if any
func (ws *WebsocketClient) SendToClient(payload user.Message) error {
	json, err := json.Marshal(&payload)
	if err != nil {
		return err
	}
	if err = ws.sock.WriteMessage(gorillaSocket.TextMessage, json); err != nil {
		return err
	}
	return nil
}

func (wc *WebsocketClient) Close() error {
	err := wc.sock.Close()
	wc.usr.State.SetWsState(user.Done)
	// update session and close if needed
	sessions.UpdateSessions()
	return err
}

// send keepalive message to client
func (ws *WebsocketClient) pong() error {
	if err := ws.sock.WriteMessage(gorillaSocket.TextMessage, []byte("pong")); err != nil {
		return err
	}
	return nil
}

// main loop
func (ws *WebsocketClient) Loop(config *configuration.Configuration, sessionID uint32) {
	// close properly
	defer ws.Close()
	// start reading from ws and write outputs to channel,
	go func() {
		for ws.usr.State.GetWsState() == user.Connected {
			_, message, err := ws.sock.ReadMessage()
			if err != nil {
				log.Printf("user %v in session %v got error from websocket: %v\n", ws.usr.Uuid, sessionID, err)
				ws.usr.State.KillWs()
				break
			}
			parsedMsg := user.Message{}
			if err := json.Unmarshal(message, &parsedMsg); err != nil {
				continue
			}
			ws.usr.WsMessageBuffer.PushToServerBuffer(parsedMsg)
		}
		log.Printf("user %v in session %v exit from ws readr\n", ws.usr.Uuid, sessionID)
	}()
	// new empty webrtc client + connection
	rtcClient, err := webrtc.NewWebrtcClient(config, ws.usr, sessionID)
	if err != nil {
		// error creating rtc, no need for sock either
		log.Printf("user %v in session %v got error creating rtc client: %v\n", ws.usr.Uuid, sessionID, err)
		return
	}
	// start rtc loop
	go rtcClient.Loop(config, sessionID)
	for {
		select {
		case <-ws.usr.State.WsKilled():
			log.Printf("user %v in session %v killed ws\n", ws.usr.Uuid, sessionID)
			return
		case <-ws.usr.State.RtcKilled():
			log.Printf("user %v in session %v killed rtc, also killing ws\n", ws.usr.Uuid, sessionID)
			return
		// send pong every few seconds
		case <-time.After(5 * time.Second):
			ws.pong()
		// requests made to the server by the server or client
		case sockMsg := <-ws.usr.WsMessageBuffer.ReadFromServerBuffer():
			switch sockMsg.Type {
			case user.MESSAGE_STARTRTC:
				// notify webrtc
				ws.usr.RtcMessageBuffer.PushToServerBuffer(user.Message{
					Type: user.MESSAGE_STARTRTC,
				})
			case user.MESSAGE_SDP:
				var sdp pwrtc.SessionDescription
				if err := json.Unmarshal(sockMsg.RawPayload, &sdp); err != nil {
					log.Printf("user %v in session %v sent bad sdp\n", ws.usr.Uuid, sessionID)
					continue
				}
				// forward to webrtc buffer
				ws.usr.RtcMessageBuffer.PushToServerBuffer(sockMsg)
			case user.MESSAGE_ICECANDIDATE:
				var icecandidate pwrtc.ICECandidate
				if err := json.Unmarshal(sockMsg.RawPayload, &icecandidate); err != nil {
					log.Printf("user %v in session %v sent bad icecandidate\n", ws.usr.Uuid, sessionID)
					continue
				}
				// forward to webrtc buffer
				ws.usr.RtcMessageBuffer.PushToServerBuffer(sockMsg)
			default:
				log.Printf("user %v in session %v got bad payload type in server\n", ws.usr.Uuid, sessionID)
			}
		// requests to make to the client
		case sockMsg := <-ws.usr.WsMessageBuffer.ReadFromClientBuffer():
			switch sockMsg.Type {
			// forward to client
			case user.MESSAGE_SDP, user.MESSAGE_ICECANDIDATE, user.MESSAGE_PCFAILED:
				ws.SendToClient(sockMsg)
			default:
				log.Printf("user %v in session %v got bad payload type in client\n", ws.usr.Uuid, sessionID)
			}
		}
	}
}
