package walle

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"time"

	"github.com/deepakkamesh/termdraw"
	"github.com/golang/glog"
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

type Emotion struct {
	term     *termdraw.Term
	emotions map[byte][]image.Image
}

func NewEmotion() *Emotion {
	return &Emotion{
		term: termdraw.New(),
	}

}

func (s *Emotion) Init(resPath string) error {

	if err := s.term.Init(); err != nil {
		return err
	}

	// Load emotions.
	exNorm, err := termdraw.LoadImages(resPath + "/walle_normal.png")
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSpeak, err := termdraw.LoadImages(
		resPath+"/walle_normal.png",
		resPath+"/walle_speaking_small.png",
		resPath+"/walle_speaking_med.png",
		resPath+"/walle_speaking_large.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exBlink, err := termdraw.LoadImages(
		resPath+"/walle_normal.png",
		resPath+"/walle_normal_eye_small.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exHappy, err := termdraw.LoadImages(
		resPath + "/walle_happy.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSad, err := termdraw.LoadImages(
		resPath + "/walle_sad.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exAngry, err := termdraw.LoadImages(
		resPath + "/walle_angry.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exPuzzled, err := termdraw.LoadImages(
		resPath + "/walle_puzzled.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exSmileMed, err := termdraw.LoadImages(
		resPath + "/walle_smile_medium.png",
	)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exThinking, err := termdraw.LoadImages(
		resPath+"/walle_normal_eyes_left.png",
		resPath+"/walle_normal_eyes_right.png",
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
	s.term.Animate(s.emotions[EMOTION_NORM], CH, 200)
	return nil
}

func (s *Emotion) Expression(emotion byte, ch rune, d int) error {
	e, ok := s.emotions[emotion]
	if !ok {
		return fmt.Errorf("expression not found")
	}
	s.term.Animate(e, CH, 200*time.Millisecond)
	return nil
}

func (s *Emotion) Quit() {
	s.term.Quit()
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
