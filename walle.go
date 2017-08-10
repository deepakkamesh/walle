package walle

import (
	"bufio"
	"os"
	"walle/assistant"
	"walle/audio"

	"github.com/golang/glog"
)

type WallEConfig struct {
	Audio      *audio.Audio
	GAssistant *assistant.GAssistant
}

type WallE struct {
	gAssistant *assistant.GAssistant // Google Assistant object.
	audio      *audio.Audio          // PortAudio IO object.
}

func New(c *WallEConfig) *WallE {

	return &WallE{
		gAssistant: c.GAssistant,
		audio:      c.Audio,
	}
}

func (s *WallE) Run() {

	s.audio.Run()
	if err := s.gAssistant.Auth(); err != nil {
		glog.Fatalf("Failed to authenticate assistant: %v", err)
	}

	for {
		audioOut := s.gAssistant.ConverseWithAssistant()

		// Convert assistant audio to text.
		txt, err := SpeechToText(audioOut)
		if err != nil {
			glog.Errorf("Failed to recognize speech: %v", err)
			continue
		}
		glog.V(1).Infof("Google Assistant said: %v", txt)

		// Get sentiment analysis of text.
		score, magnitude, err := AnalyzeSentiment(txt)
		if err != nil {
			glog.Errorf("Failed to analyze sentiment: %v", err)
			continue
		}
		glog.V(1).Infof("Sentiment Analysis - Score:%v Magnitude:%v", score, magnitude)

		// Wait for enter.
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
	}
}
