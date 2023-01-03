package user

import (
	"pion-webrtc-sfu/configuration"
)

// user preferences and message buffers
type User struct {
	Uuid             string        // unique user id
	Settings         userSettings  // user settings
	State            userState     // user activity status,
	WsMessageBuffer  messageBuffer // client-ws message buffer
	RtcMessageBuffer messageBuffer // client-rtc message buffer
}

func NewUser(uuid string, config *configuration.Configuration) User {
	return User{
		Uuid:             uuid,
		Settings:         newUserSettings(config.Rtc_disconnect_timeout_seconds, config.Rtc_failed_timeout_seconds, config.Rtc_keepalive_interval_seconds),
		State:            newUserState(),
		WsMessageBuffer:  NewMessageBuffer(),
		RtcMessageBuffer: NewMessageBuffer(),
	}
}
