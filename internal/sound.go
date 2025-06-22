package internal

import (
	"embed"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"time"
)

type SoundPlayer struct {
	fs    embed.FS
	logCh chan<- LogMessage
}

func NewSoundPlayer(logCh chan<- LogMessage, fs embed.FS) *SoundPlayer {
	return &SoundPlayer{
		fs:    fs,
		logCh: logCh,
	}
}

func (s *SoundPlayer) PlaySound(sound string) {
	f, err := s.fs.Open(fmt.Sprintf("res/%s.wav", sound))
	if err != nil {
		s.logCh <- LogMessage{LogSeverityFail, "failed to open sound file: " + err.Error()}
		return
	}

	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		s.logCh <- LogMessage{LogSeverityFail, "failed to decode sound file: " + err.Error()}
		return
	}

	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		s.logCh <- LogMessage{LogSeverityFail, "failed to init speaker: " + err.Error()}
		return
	}

	volume := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   -3.5,
		Silent:   false,
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(volume, beep.Callback(func() {
		done <- true
	})))

	<-done
}
