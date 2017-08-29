package walle

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"time"

	"github.com/deepakkamesh/termdraw"
	"github.com/golang/glog"
	"gobot.io/x/gobot/platforms/raspi"
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

// Face represents a struct making up the moving parts.
type Face struct {
	leftEye  []image.Image
	rightEye []image.Image
}

type Emotion struct {
	term         *termdraw.Term
	leftEye      *OLED
	rightEye     *OLED
	termEmotions map[byte][]image.Image
	faceEmotions map[byte]Face
}

func NewEmotion() *Emotion {
	return &Emotion{
		term:     termdraw.New(),
		leftEye:  NewOLED(),
		rightEye: NewOLED(),
	}
}

func (s *Emotion) Init(resPath string) error {

	if err := s.term.Init(); err != nil {
		return err
	}

	// Initialize Pi Adapter.
	r := raspi.NewAdaptor()
	if err := r.Connect(); err != nil {
		return err
	}

	if err := s.leftEye.Init(r, 1, 0x3c); err != nil {
		return err
	}

	if err := s.rightEye.Init(r, 1, 0x3d); err != nil {
		return err
	}

	// Start up "facial muscles".
	if err := s.term.Run(); err != nil {
		return err
	}
	if err := s.leftEye.Run(); err != nil {
		return err
	}

	if err := s.rightEye.Run(); err != nil {
		return err
	}

	// Load images from disk.
	termEmotions, err := loadTermEmotions(resPath)
	if err != nil {
		return err
	}
	s.termEmotions = termEmotions

	faceEmotions, err := loadFaceEmotion(resPath)
	if err != nil {
		return err
	}
	s.faceEmotions = faceEmotions

	// Default expression.
	s.term.Animate(s.termEmotions[EMOTION_NORM], CH, 200)
	s.leftEye.Animate(s.faceEmotions[EMOTION_NORM].leftEye, 100)
	s.rightEye.Animate(s.faceEmotions[EMOTION_NORM].rightEye, 100)

	return nil
}

// Expression displays the requested emotion using the character ch. If the expression is
// animated it switchesusing ms milliseconds.
func (s *Emotion) Expression(emotion byte, ch rune, ms uint16) error {
	e, ok := s.termEmotions[emotion]
	if !ok {
		return fmt.Errorf("expression not found")
	}
	s.term.Animate(e, CH, 200*time.Millisecond)

	face, ok := s.faceEmotions[emotion]
	if !ok {
		return fmt.Errorf("expression not found")
	}

	switch emotion {
	case EMOTION_NORM:
		s.leftEye.Animate(face.leftEye, 100*time.Millisecond)
		s.rightEye.Animate(face.rightEye, 100*time.Millisecond)
	}

	return nil
}

func (s *Emotion) Quit() {
	s.term.Quit()
	s.leftEye.Quit()
	s.rightEye.Quit()
}

// selectEmotion allows overriding of emotions based on txt.
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

func loadTermEmotions(resPath string) (map[byte][]image.Image, error) {
	// Load emotions.
	exNorm, err := termdraw.LoadImages(resPath + "/walle_normal.png")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSpeak, err := termdraw.LoadImages(
		resPath+"/walle_normal.png",
		resPath+"/walle_speaking_small.png",
		resPath+"/walle_speaking_med.png",
		resPath+"/walle_speaking_large.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exBlink, err := termdraw.LoadImages(
		resPath+"/walle_normal.png",
		resPath+"/walle_normal_eye_small.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exHappy, err := termdraw.LoadImages(
		resPath + "/walle_happy.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	exSad, err := termdraw.LoadImages(
		resPath + "/walle_sad.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exAngry, err := termdraw.LoadImages(
		resPath + "/walle_angry.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exPuzzled, err := termdraw.LoadImages(
		resPath + "/walle_puzzled.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exSmileMed, err := termdraw.LoadImages(
		resPath + "/walle_smile_medium.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	exThinking, err := termdraw.LoadImages(
		resPath+"/walle_normal_eyes_left.png",
		resPath+"/walle_normal_eyes_right.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	return map[byte][]image.Image{
		EMOTION_NORM:      exNorm,
		EMOTION_SPEAK:     exSpeak,
		EMOTION_BLINK:     exBlink,
		EMOTION_HAPPY:     exHappy,
		EMOTION_ANGRY:     exAngry,
		EMOTION_SAD:       exSad,
		EMOTION_PUZZLED:   exPuzzled,
		EMOTION_SMILE_MED: exSmileMed,
		EMOTION_THINKING:  exThinking,
	}, nil
}

func loadFaceEmotion(resPath string) (map[byte]Face, error) {

	leftINorm, err := LoadImages(
		resPath + "/left_eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}
	rightINorm, err := LoadImages(
		resPath + "/right_eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	return map[byte]Face{
		EMOTION_NORM:      Face{leftINorm, rightINorm},
		EMOTION_SPEAK:     Face{},
		EMOTION_BLINK:     Face{},
		EMOTION_HAPPY:     Face{},
		EMOTION_ANGRY:     Face{},
		EMOTION_SAD:       Face{},
		EMOTION_PUZZLED:   Face{},
		EMOTION_SMILE_MED: Face{},
		EMOTION_THINKING:  Face{},
	}, nil

}
