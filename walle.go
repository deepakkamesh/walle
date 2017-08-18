package walle

import (
	"errors"
	"fmt"
	"image"

	embedded "google.golang.org/genproto/googleapis/assistant/embedded/v1alpha1"

	"github.com/deepakkamesh/termdraw"
	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
	termbox "github.com/nsf/termbox-go"
)

const (
	EMOTION_NORM byte = iota
	EMOTION_SPEAK
)

type WallEConfig struct {
	Audio        *audio.Audio
	GAssistant   *assistant.GAssistant
	Term         *termdraw.Term
	ResourcePath string
}

type WallE struct {
	audio      *audio.Audio          // PortAudio IO object.
	gAssistant *assistant.GAssistant // Google Assistant object.
	term       *termdraw.Term        // Termdraw object
	resPath    string
	emotions   map[byte][]image.Image
}

func New(c *WallEConfig) *WallE {

	return &WallE{
		gAssistant: c.GAssistant,
		audio:      c.Audio,
		term:       c.Term,
		resPath:    c.ResourcePath,
		emotions:   make(map[byte][]image.Image),
	}
}

func (s *WallE) Init() error {
	s.audio.StartPlayback()

	if err := s.term.Init(); err != nil {
		return err
	}
	if err := s.gAssistant.Auth(); err != nil {
		return err
	}

	//	 Load emotions.
	exNorm, err := termdraw.LoadImages(s.resPath + "/walle_normal.png")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSpeak, err := termdraw.LoadImages(s.resPath+"/walle_normal.png",
		s.resPath+"/walle_speaking_small.png", s.resPath+"/walle_speaking_med.png",
		s.resPath+"/walle_speaking_large.png")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	s.emotions[EMOTION_NORM] = exNorm
	s.emotions[EMOTION_SPEAK] = exSpeak
	s.term.Run()

	return nil
}

func (s *WallE) Run() {
	for {
		evt := <-s.term.EventCh
		if evt.Type == termbox.EventKey {
			switch {
			case evt.Key == termbox.KeyEsc:
				s.term.Quit()
				return
			case evt.Ch == 'r':
				s.interact()
			}

		}
	}
	return
}
func (s *WallE) interact() {

	go func() {
		st := <-s.gAssistant.StatusCh
		if st == embedded.ConverseResponse_END_OF_UTTERANCE {
			glog.V(2).Infof("END_UTTERNACE")
			s.term.Animate(s.emotions[EMOTION_SPEAK], '*', 200)
		}
	}()

	s.term.Animate(s.emotions[EMOTION_NORM], '*', 200)
	audioOut := s.gAssistant.ConverseWithAssistant()
	s.term.Animate(s.emotions[EMOTION_NORM], '*', 200)

	// Convert assistant audio to text.
	txt, err := SpeechToText(audioOut)
	if err != nil {
		glog.Errorf("Failed to recognize speech: %v", err)
		return
	}
	glog.V(1).Infof("Google Assistant said: %v", txt)

	// Get sentiment analysis of text.
	score, magnitude, err := AnalyzeSentiment(txt)
	if err != nil {
		glog.Errorf("Failed to analyze sentiment: %v", err)
		return
	}
	glog.V(1).Infof("Sentiment Analysis - Score:%v Magnitude:%v", score, magnitude)

}
