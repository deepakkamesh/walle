package walle

import (
	"bufio"
	"os"

	embedded "google.golang.org/genproto/googleapis/assistant/embedded/v1alpha1"

	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
)

type WallEConfig struct {
	Audio      *audio.Audio
	GAssistant *assistant.GAssistant
}

type WallE struct {
	audio      *audio.Audio          // PortAudio IO object.
	gAssistant *assistant.GAssistant // Google Assistant object.
}

func New(c *WallEConfig) *WallE {

	return &WallE{
		gAssistant: c.GAssistant,
		audio:      c.Audio,
	}
}

func (s *WallE) Run() {

	s.audio.StartPlayback()

	if err := s.gAssistant.Auth(); err != nil {
		glog.Fatalf("Failed to authenticate assistant: %v", err)
	}

	for {

		go func() {
			st := <-s.gAssistant.StatusCh
			if st == embedded.ConverseResponse_END_OF_UTTERANCE {
				glog.V(2).Infof("END_UTTERNACE")
			}
		}()

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
