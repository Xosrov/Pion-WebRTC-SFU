package http

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pion-webrtc-sfu/configuration"
	"pion-webrtc-sfu/sessions"
	"pion-webrtc-sfu/tracks"
	"pion-webrtc-sfu/user"
	"pion-webrtc-sfu/websocket"
)

// start http listener
func ServeHttp(config *configuration.Configuration) {
	// set mode
	if config.Http_gin_is_debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	is_ssl := config.Http_tls_cert_file_location != "" && config.Http_tls_key_file_location != ""

	// html test server enabled
	if config.Http_local_htmlserver_enabled {
		// location of static files
		router.Static("/static", "http/view/static")
		router.LoadHTMLGlob("http/view/html/*")
		router.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
	}

	// websocket server - always served
	router.GET("/ws", func(c *gin.Context) {
		// get session ID and user ID from the user.
		// this might need to be changed later since
		// these values might come from an external API.
		sid := c.Request.URL.Query().Get("sid")    // unique for every user
		userID := c.Request.URL.Query().Get("uid") // unique for every streaming session (can contain multiple users)
		if sid == "" || userID == "" {
			log.Printf("connection attempted with no sid and uid provided, declining\n")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		// parse sessionID - uint32 to be
		sessionID, err := strconv.ParseUint(sid, 10, 32)
		if err != nil {
			log.Printf("user %v did not provide a valid ssrc!, declining connection!\n", userID)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// create user with id
		u := user.NewUser(userID, config)
		// create track group
		trackGroup, err := tracks.NewTrackGroup(config)
		if err != nil {
			log.Printf("session %v, user %v: server could not create track group\n", sessionID, userID)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// add user to session
		sess := sessions.AddSession(uint32(sessionID), trackGroup)
		// add user to session
		if err := sess.AddUser(&u); err != nil {
			log.Printf("session %v, user %v: user already exists in session\n", sessionID, userID)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		// create websocket client
		wsClient, err := websocket.NewWebsocketClient(c, &u)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Printf("session %v, user %v could not create ws client: %v\n", sessionID, userID, err)
			// close ws
			wsClient.Close()
			return
		}
		// start websocket handler loop
		go wsClient.Loop(config, uint32(sessionID))
	})
	// serve with ssl if specified
	if is_ssl {
		go router.RunTLS(config.Http_local_server_location, config.Http_tls_cert_file_location, config.Http_tls_key_file_location)
	} else {
		go router.Run(config.Http_local_server_location)
	}
}
