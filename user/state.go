package user

import (
	"sync"
)

// status codes
type State int

// defined states
const (
	Disconnected State = iota
	AwaitingConnection
	Connected
	Done
	killed // only used internally
)

// TODO: in case user rtc/sock wants to reconnect after being killed or stopped or whatever ; need to reset boolean values and re open chans
// user state struct storing webrtc and websocket states of the user
type userState struct {
	wsState   State        // websocket status
	rtcState  State        // webrtc status
	wsMu      sync.RWMutex // mutex for accessing ws status
	rtcMu     sync.RWMutex // mutex for accessing rtc status
	wsKilled  chan bool    // channel to listen for ws killed state
	rtcKilled chan bool    // channel to listen for rtc killed state
}

func newUserState() userState {
	return userState{
		wsState:   Disconnected,
		rtcState:  Disconnected,
		wsKilled:  make(chan bool),
		rtcKilled: make(chan bool),
	}
}

func (us *userState) SetWsState(state State) {
	us.wsMu.Lock()
	defer us.wsMu.Unlock()
	us.wsState = state
}

func (us *userState) GetWsState() State {
	us.wsMu.RLock()
	defer us.wsMu.RUnlock()
	return us.wsState
}

func (us *userState) WsKilled() <-chan bool {
	return us.wsKilled
}

func (us *userState) SetRtcState(state State) {
	us.rtcMu.Lock()
	defer us.rtcMu.Unlock()
	us.rtcState = state
}

func (us *userState) GetRtcState() State {
	us.rtcMu.RLock()
	defer us.rtcMu.RUnlock()
	return us.rtcState
}

func (us *userState) RtcKilled() <-chan bool {
	return us.rtcKilled
}

// closes ws chan and updates state
func (us *userState) KillWs() {
	us.wsMu.Lock()
	defer us.wsMu.Unlock()
	// already killed, ignoring
	if us.wsState == killed {
		return
	}
	// close this channel, which allows it to be read from indefinitely
	close(us.wsKilled)
	us.wsState = killed
}

// closes rtc chan and updates state
func (us *userState) KillRtc() {
	us.rtcMu.Lock()
	defer us.rtcMu.Unlock()
	// already killed, ignoring
	if us.rtcState == killed {
		return
	}
	// close this channel, which allows it to be read from indefinitely
	close(us.rtcKilled)
	us.rtcState = killed
}
