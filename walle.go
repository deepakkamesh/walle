package walle

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"time"

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
	EMOTION_BLINK
	EMOTION_HAPPY
	EMOTION_ANGRY
	EMOTION_SAD
	EMOTION_PUZZLED
	EMOTION_SMILE_MED
	EMOTION_THINKING
)
const (
	T500 = 500 * time.Millisecond
	T300 = 300 * time.Millisecond
	T200 = 200 * time.Millisecond
	T100 = 100 * time.Millisecond
)
const (
	CH1 = '█'
	CH  = '▒'
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
	term       *termdraw.Term        // Termdraw object.
	resPath    string
	emotions   map[byte][]image.Image
}

// New returns a new initialized WallE object.
func New(c *WallEConfig) *WallE {

	return &WallE{
		gAssistant: c.GAssistant,
		audio:      c.Audio,
		term:       c.Term,
		resPath:    c.ResourcePath,
		//emotions:   make(map[byte][]image.Image), // TODO: Remove?
	}
}

// Init initializes WallE subsystems (gAssistant, Audio, termdraw).
func (s *WallE) Init() error {
	s.audio.StartPlayback()
	if err := s.term.Init(); err != nil {
		return err
	}
	if err := s.gAssistant.Auth(); err != nil {
		return err
	}

	// Load emotions.
	exNorm, err := termdraw.LoadImages(s.resPath + "/walle_normal.png")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSpeak, err := termdraw.LoadImages(
		s.resPath+"/walle_normal.png",
		s.resPath+"/walle_speaking_small.png",
		s.resPath+"/walle_speaking_med.png",
		s.resPath+"/walle_speaking_large.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exBlink, err := termdraw.LoadImages(
		s.resPath+"/walle_normal.png",
		s.resPath+"/walle_normal_eye_small.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exHappy, err := termdraw.LoadImages(
		s.resPath + "/walle_happy.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSad, err := termdraw.LoadImages(
		s.resPath + "/walle_sad.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exAngry, err := termdraw.LoadImages(
		s.resPath + "/walle_angry.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exPuzzled, err := termdraw.LoadImages(
		s.resPath + "/walle_puzzled.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exSmileMed, err := termdraw.LoadImages(
		s.resPath + "/walle_smile_medium.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exThinking, err := termdraw.LoadImages(
		s.resPath+"/walle_normal_eyes_left.png",
		s.resPath+"/walle_normal_eyes_right.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	s.emotions = map[byte][]image.Image{
		EMOTION_NORM:      exNorm,
		EMOTION_SPEAK:     exSpeak,
		EMOTION_BLINK:     exBlink,
		EMOTION_HAPPY:     exHappy,
		EMOTION_ANGRY:     exAngry,
		EMOTION_SAD:       exSad,
		EMOTION_PUZZLED:   exPuzzled,
		EMOTION_SMILE_MED: exSmileMed,
		EMOTION_THINKING:  exThinking,
	}

	s.term.Run()
	s.term.Animate(s.emotions[EMOTION_NORM], CH, T200)

	return nil
}

// Run is the main event loop.
func (s *WallE) Run() {
	for {
		evt := <-s.term.EventCh
		if evt.Type == termbox.EventKey {
			switch {
			case evt.Key == termbox.KeyEsc:
				s.term.Quit()
				s.audio.Quit()
				return
			case evt.Ch == 'r':
				s.interactAI()
			}
		}
	}
	return
}

// interactAI runs a gAssistant session collects the response text
// and analyzes it for sentiment.
func (s *WallE) interactAI() {

	//TODO: This is a workaround for Pi as the audio does not continue playing. Needs
	// investigation and fix.
	s.audio.ResetPlayback()

	go func() {
		st := <-s.gAssistant.StatusCh
		if st == embedded.ConverseResponse_END_OF_UTTERANCE {
			glog.V(2).Infof("GAssisant sent END_OF_UTTERNACE")
			s.term.Animate(s.emotions[EMOTION_SPEAK], CH, T100)
		}
	}()

	s.term.Animate(s.emotions[EMOTION_BLINK], CH, T200)
	audioOut := s.gAssistant.ConverseWithAssistant()
	s.term.Animate(s.emotions[EMOTION_THINKING], CH, T500)

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
	s.term.Animate(s.emotions[emotion], CH, T200)
}

func selectEmotion(score float32, txt string) byte {

	// Override sentiment.
	words := map[byte][]string{
		EMOTION_SAD:   {"sorry", "apologies"},
		EMOTION_HAPPY: {"joke", "laugh"},
	}

	for emotion, wordList := range words {
		for _, word := range wordList {
			if strings.Contains(txt, word) {
				glog.V(2).Infof("Emotion override for matching word %v", word)
				return emotion
			}
		}
	}
	switch {
	case 0.4 < score && score < 1:
		return EMOTION_HAPPY
	case 0.1 < score && score <= 0.4:
		return EMOTION_SMILE_MED
	case -0.2 < score && score <= 0.1:
		return EMOTION_NORM
	case -0.6 < score && score <= -0.2:
		return EMOTION_SAD
	case -1 < score && score <= -0.6:
		return EMOTION_ANGRY
	}
	return EMOTION_NORM
}
