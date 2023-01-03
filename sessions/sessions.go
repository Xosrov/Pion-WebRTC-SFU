package sessions

import (
	"errors"
	"pion-webrtc-sfu/tracks"
	"pion-webrtc-sfu/user"
	"sync"
)

type Session struct {
	TrackGroup     tracks.TrackGroup // RTP track shared between all users
	ConnectedUsers []*user.User      // connected users
	sync.RWMutex                     // mutex for user list read/write
}

/*
session manager for http / sock / rtc
map unique session ID to a session
ID used is the ssrc of the incoming rtp stream
*/
var sessions map[uint32]*Session

// mutex for above map read/write
var mutex sync.Mutex

// initiate session map
func InitSessions() {
	mutex.Lock()
	defer mutex.Unlock()
	sessions = make(map[uint32]*Session)
}

// add user to a session
func (s *Session) AddUser(usr *user.User) error {
	s.RWMutex.Lock()
	defer s.RWMutex.Unlock()
	// check if user exists already
	for _, u := range s.ConnectedUsers {
		if u.Uuid == usr.Uuid {
			return errors.New("user already exists")
		}
	}
	s.ConnectedUsers = append(s.ConnectedUsers, usr)
	return nil
}

/*
check if remote IP exists in current sessions
return session if it does
*/
func ReturnSessionByIdIfExists(id uint32) *Session {
	mutex.Lock()
	defer mutex.Unlock()
	session, ok := sessions[id]
	if ok {
		return session
	}
	return nil
}

// add a new session with id if it doesn't exist and return in
func AddSession(id uint32, trackGroup tracks.TrackGroup) *Session {
	mutex.Lock()
	defer mutex.Unlock()
	old, exists := sessions[id]
	// already exists, return old
	if exists {
		return old
	}
	new := &Session{
		TrackGroup:     trackGroup,
		ConnectedUsers: make([]*user.User, 0),
	}
	sessions[id] = new
	return new
}

/*
update sessions (remove inactive users)
run this once in a while or when user is disconnected
*/
func UpdateSessions() {
	mutex.Lock()
	defer mutex.Unlock()
	for is, session := range sessions {
		for iu, u := range session.ConnectedUsers {
			// if no activity
			if u.State.GetRtcState() == user.Done && u.State.GetWsState() == user.Done {
				// remove from users
				session.ConnectedUsers = append(session.ConnectedUsers[:iu], session.ConnectedUsers[iu+1:]...)
			}
		}
		// delete session if no more users in it
		if len(session.ConnectedUsers) == 0 {
			delete(sessions, is)
		}
	}
}
