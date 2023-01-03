package main

import (
	"fmt"
	"log"
	"pion-webrtc-sfu/configuration"
	"pion-webrtc-sfu/http"
	"pion-webrtc-sfu/sessions"
	"pion-webrtc-sfu/writer"
	"runtime"
	"time"
)

func main() {
	// set logger flags
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile)
	// load configuration from file
	conf, err := configuration.CreateConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	configuration.PrintConfiguration(conf)
	// initiate empty sessions
	sessions.InitSessions()
	go writer.StartVideoWriterLoop(conf)
	go writer.StartAudioWriterLoop(conf)
	go http.ServeHttp(conf)
	var m runtime.MemStats
	for {
		runtime.ReadMemStats(&m)
		log.SetPrefix("\r") // to override goroutine counter if it exists
		fmt.Printf("\rNumber of goroutines: %d, Alloc = %vB, TAlloc = %vB, Sys = %vB", runtime.NumGoroutine(), m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024)
		time.Sleep(time.Second * 15)
	}
}
