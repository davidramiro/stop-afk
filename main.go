package main

import (
	"stop-afk/internal"
)

const (
	version = "1.0.0"
	port    = 4242
)

func main() {
	logCh := make(chan internal.LogMessage)
	roundCh := make(chan internal.Round)

	u := internal.NewUI(logCh, roundCh)
	go u.Start()
	go u.ProcessChannels()

	c := internal.NewConfig(logCh)
	go c.Init(version, port)

	s := internal.NewServer(port, roundCh)

	go s.StartListener()

	logCh <- internal.LogMessage{Severity: internal.LogSeverityOK, Message: "started gamestate listener"}
	select {}
}
