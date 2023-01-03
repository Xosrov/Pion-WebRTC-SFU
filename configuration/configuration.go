package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type PortRange struct {
	Start uint16
	End   uint16
}

type Configuration struct {
	Http_local_server_location      string
	Http_local_htmlserver_enabled   bool
	Http_tls_cert_file_location     string
	Http_tls_key_file_location      string
	Http_gin_is_debug               bool
	Rtc_disconnect_timeout_seconds  uint
	Rtc_video_tracks_receive_port   uint16
	Rtc_audio_tracks_receive_port   uint16
	Rtc_receive_rtp_buffsize        uint16
	Rtc_video_codec                 string
	Rtc_audio_codec                 string
	Rtc_failed_timeout_seconds      uint
	Rtc_keepalive_interval_seconds  uint
	Server_ephemeral_udp_port_range PortRange
	Server_NAT_1to1_IPs             string
}

func PrintConfiguration(config *Configuration) {
	s, _ := json.MarshalIndent(config, "", "\t")
	log.Printf("Configuration: \n%s", string(s))
}

// get value of env_key from loaded .env and return it
// the returned type is same as default_value's type
func valueFromEnv(env_key string, default_value interface{}) (interface{}, error) {
	env_value_str := os.Getenv(env_key)
	// use default value if nothing set in .env
	if env_value_str == "" {
		return default_value, nil
	}
	switch default_value.(type) {
	case int:
		value, err := strconv.Atoi(env_value_str)
		if err != nil {
			return nil, err
		}
		return value, nil
	case uint:
		value, err := strconv.ParseUint(env_value_str, 10, 0)
		if err != nil {
			return nil, err
		}
		return uint(value), nil
	case uint16:
		value, err := strconv.ParseUint(env_value_str, 10, 16)
		if err != nil {
			return nil, err
		}
		return uint16(value), nil
	case string:
		return env_value_str, nil
	case bool:
		env_value_str = strings.ToLower(env_value_str)
		if env_value_str == "true" {
			return true, nil
		} else if env_value_str == "false" {
			return false, nil
		}
		return nil, errors.New("unknown value type (allowed values are true and false)")
	case PortRange:
		split := strings.SplitN(env_value_str, "-", 2)
		if len(split) != 2 {
			return nil, errors.New("port range needs to be in the format MIN-MAX")
		}
		minval, err := strconv.ParseUint(split[0], 10, 16)
		if err != nil {
			return nil, errors.New("invalid port range value types (need uint16)")
		}
		maxval, err := strconv.ParseUint(split[1], 10, 16)
		if err != nil {
			return nil, errors.New("invalid port range value types (need uint16)")
		}
		// reverse order, fix it
		if minval > maxval {
			temp := minval
			minval = maxval
			maxval = temp
		}
		// convert values to uint16
		return PortRange{uint16(minval), uint16(maxval)}, nil
	}
	return nil, errors.New("unknown type to read from env")
}

// read values from .env and return configuration
func CreateConfiguration() (*Configuration, error) {
	// load .env config
	godotenv.Load(ENV_FILE)

	http_local_server_location, err := valueFromEnv("HTTP_LOCAL_SERVER_LOCATION", HTTP_LOCAL_SERVER_LOCATION_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP_LOCAL_SERVER_LOCATION: %v", err)
	}
	http_local_htmlserver_enabled, err := valueFromEnv("HTTP_LOCAL_HTMLSERVER_ENABLED", HTTP_LOCAL_HTMLSERVER_ENABLED_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP_LOCAL_HTMLSERVER_ENABLED: %v", err)
	}
	http_tls_cert_file_location, err := valueFromEnv("HTTP_TLS_CERT_FILE_LOCATION", HTTP_TLS_CERT_FILE_LOCATION_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP_TLS_CERT_FILE_LOCATION: %v", err)
	}
	http_tls_key_file_location, err := valueFromEnv("HTTP_TLS_KEY_FILE_LOCATION", HTTP_TLS_KEY_FILE_LOCATION_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP_TLS_KEY_FILE_LOCATION: %v", err)
	}
	http_gin_is_debug, err := valueFromEnv("HTTP_GIN_IS_DEBUG", HTTP_GIN_IS_DEBUG_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading HTTP_GIN_IS_DEBUG: %v", err)
	}
	rtc_video_tracks_receive_port, err := valueFromEnv("RTC_VIDEO_TRACKS_RECEIVE_PORT", RTC_VIDEO_TRACKS_RECEIVE_PORT_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_VIDEO_TRACKS_RECEIVE_PORT: %v", err)
	}
	rtc_audio_tracks_receive_port, err := valueFromEnv("RTC_AUDIO_TRACKS_RECEIVE_PORT", RTC_AUDIO_TRACKS_RECEIVE_PORT_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_AUDIO_TRACKS_RECEIVE_PORT: %v", err)
	}
	rtc_receive_rtp_buffsize, err := valueFromEnv("RTC_RECEIVE_RTP_BUFFSIZE", RTC_RECEIVE_RTP_BUFFSIZE_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_RECEIVE_RTP_BUFFSIZE: %v", err)
	}
	rtc_video_codec, err := valueFromEnv("RTC_VIDEO_CODEC", RTC_VIDEO_CODEC_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_VIDEO_CODEC: %v", err)
	}
	rtc_audio_codec, err := valueFromEnv("RTC_AUDIO_CODEC", RTC_AUDIO_CODEC_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_AUDIO_CODEC: %v", err)
	}
	rtc_disconnect_timeout_seconds, err := valueFromEnv("RTC_DISCONNECT_TIMEOUT_SECONDS", RTC_DISCONNECT_TIMEOUT_SECONDS_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_DISCONNECT_TIMEOUT_SECONDS: %v", err)
	}
	rtc_failed_timeout_seconds, err := valueFromEnv("RTC_FAILED_TIMEOUT_SECONDS", RTC_FAILED_TIMEOUT_SECONDS_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_FAILED_TIMEOUT_SECONDS: %v", err)
	}
	rtc_keepalive_interval_seconds, err := valueFromEnv("RTC_KEEPALIVE_INTERVAL_SECONDS", RTC_KEEPALIVE_INTERVAL_SECONDS_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading RTC_KEEPALIVE_INTERVAL_SECONDS: %v", err)
	}
	server_ephemeral_udp_port_range, err := valueFromEnv("SERVER_EPHEMERAL_UDP_PORT_RANGE", PortRange{SERVER_EPHEMERAL_UDP_PORT_RANGE_MIN_DEFAULT, SERVER_EPHEMERAL_UDP_PORT_RANGE_MAX_DEFAULT})
	if err != nil {
		return nil, fmt.Errorf("error reading SERVER_EPHEMERAL_UDP_PORT_RANGE: %v", err)
	}
	server_nat_1to1_ips, err := valueFromEnv("SERVER_NAT_1TO1_IPS", SERVER_NAT_1TO1_IPS_DEFAULT)
	if err != nil {
		return nil, fmt.Errorf("error reading SERVER_NAT_1TO1_IP: %v", err)
	}

	return &Configuration{
		Http_local_server_location:      http_local_server_location.(string),
		Http_local_htmlserver_enabled:   http_local_htmlserver_enabled.(bool),
		Http_tls_cert_file_location:     http_tls_cert_file_location.(string),
		Http_tls_key_file_location:      http_tls_key_file_location.(string),
		Http_gin_is_debug:               http_gin_is_debug.(bool),
		Rtc_video_tracks_receive_port:   rtc_video_tracks_receive_port.(uint16),
		Rtc_audio_tracks_receive_port:   rtc_audio_tracks_receive_port.(uint16),
		Rtc_receive_rtp_buffsize:        rtc_receive_rtp_buffsize.(uint16),
		Rtc_video_codec:                 rtc_video_codec.(string),
		Rtc_audio_codec:                 rtc_audio_codec.(string),
		Rtc_disconnect_timeout_seconds:  rtc_disconnect_timeout_seconds.(uint),
		Rtc_failed_timeout_seconds:      rtc_failed_timeout_seconds.(uint),
		Rtc_keepalive_interval_seconds:  rtc_keepalive_interval_seconds.(uint),
		Server_ephemeral_udp_port_range: server_ephemeral_udp_port_range.(PortRange),
		Server_NAT_1to1_IPs:             server_nat_1to1_ips.(string),
	}, nil
}
