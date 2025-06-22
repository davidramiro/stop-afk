package main

import (
	"embed"
	_ "embed"
	"stop-afk/internal"
)

const (
	version = "1.0.1"
	port    = 4242
)

//go:embed res/buy.wav res/start.wav
var fs embed.FS

func main() {
	logCh := make(chan internal.LogMessage)
	roundCh := make(chan internal.Round)

	sp := internal.NewSoundPlayer(logCh, fs)
	u := internal.NewUI(logCh, roundCh, sp)
	go u.Start()
	go u.ProcessChannels()

	c := internal.NewConfig(logCh)
	go c.Init(version, port)

	s := internal.NewServer(port, roundCh, logCh)

	go s.StartListener()

	logCh <- internal.LogMessage{Severity: internal.LogSeverityOK, Message: "started gamestate listener"}
	select {}
}
