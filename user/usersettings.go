package user

import "time"

/*
settings for each user
can be used to construct rtc/sock
*/
type userSettings struct {
	RTCDisconnectTimeout time.Duration
	RTCFailedTimeout     time.Duration
	RTCKeepaliveInterval time.Duration
}

func newUserSettings(RTCDisconnectTimeoutSeconds, RTCFailedTimeoutSeconds, RTCKeepaliveIntervalSeconds uint) userSettings {
	return userSettings{
		RTCDisconnectTimeout: time.Second * time.Duration(RTCDisconnectTimeoutSeconds),
		RTCFailedTimeout:     time.Second * time.Duration(RTCFailedTimeoutSeconds),
		RTCKeepaliveInterval: time.Second * time.Duration(RTCKeepaliveIntervalSeconds),
	}
}
