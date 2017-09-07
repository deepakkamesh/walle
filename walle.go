package walle

import (
	"fmt"

	"gobot.io/x/gobot/platforms/raspi"
	embedded "google.golang.org/genproto/googleapis/assistant/embedded/v1alpha1"

	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
	termbox "github.com/nsf/termbox-go"
)

const (
	CH1 = '█'
	CH  = '▒'
)

type WallEConfig struct {
	Audio        *audio.Audio
	GAssistant   *assistant.GAssistant
	ResourcePath string
}

type WallE struct {
	audio      *audio.Audio          // PortAudio IO object.
	gAssistant *assistant.GAssistant // Google Assistant object.
	emotion    *Emotion
	resPath    string
}

// New returns a new initialized WallE object.
func New(c *WallEConfig) *WallE {

	return &WallE{
		gAssistant: c.GAssistant,
		audio:      c.Audio,
		emotion:    NewEmotion(),
		resPath:    c.ResourcePath,
	}
}

// Init initializes WallE subsystems (gAssistant, Audio, ).
func (s *WallE) Init() error {
	s.audio.StartPlayback()
	if err := s.gAssistant.Auth(); err != nil {
		return err
	}

	// Initialize Pi Adapter.
	rpi := raspi.NewAdaptor()
	if err := rpi.Connect(); err != nil {
		return err
	}

	if err := s.emotion.Init(s.resPath, rpi); err != nil {
		return fmt.Errorf("failed to init emotions:%v", err)
	}

	return nil
}

// Run is the main event loop.
func (s *WallE) Run() {
	for {
		evt := <-s.emotion.term.EventCh
		if evt.Type == termbox.EventKey {
			switch {
			case evt.Key == termbox.KeyEsc:
				s.emotion.Quit()
				s.audio.Quit()
				return
			case evt.Ch == 'r':
				s.interactAI()

			case evt.Ch == 't':
				s.emotion.CycleEmotions()
			}
		}
	}
	return
}

// interactAI runs a gAssistant session collects the response text
// and analyzes it for sentiment.
func (s *WallE) interactAI() {

	//TODO: This is a workaround for Pi as the audio does not continue playing after first interaction. Needs
	// investigation and fix.
	s.audio.ResetPlayback()

	go func() {
		st := <-s.gAssistant.StatusCh
		if st == embedded.ConverseResponse_END_OF_UTTERANCE {
			glog.V(2).Infof("GAssisant sent END_OF_UTTERNACE")
			s.emotion.Expression(EMOTION_SPEAK, CH, 200)
		}
	}()

	s.emotion.Expression(EMOTION_BLINK, CH, 200)
	audioOut := s.gAssistant.ConverseWithAssistant()
	s.emotion.Expression(EMOTION_THINKING, CH, 100)

	// Convert assistant audio to text.
	txt, err := SpeechToText(audioOut)
	if err != nil {
		glog.Errorf("Failed to recognize speech: %v", err)
		return //TODO: returns should change emotion.
	}
	glog.V(1).Infof("Google Assistant said: %v", txt)

	// Get sentiment analysis of text.
	score, magnitude, err := AnalyzeSentiment(txt)
	if err != nil {
		glog.Errorf("Failed to analyze sentiment: %v", err)
		return
	}
	glog.V(1).Infof("Sentiment Analysis - Score:%v Magnitude:%v", score, magnitude)

	// Select an emotion to display.
	emotion := selectEmotion(score, txt)
	s.emotion.Expression(emotion, CH, 200)
}
