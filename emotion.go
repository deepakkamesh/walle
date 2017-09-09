package walle

import (
	"errors"
	"fmt"
	"image"
	"strings"
	"sync"
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
	eye   []image.Image
	mouth []image.Image
}

type Emotion struct {
	term         *termdraw.Term
	eye          *OLED
	mouth        *OLED
	termEmotions map[byte][]image.Image
	faceEmotions map[byte]Face
}

func NewEmotion() *Emotion {
	return &Emotion{
		term:  termdraw.New(),
		eye:   NewOLED(),
		mouth: NewOLED(),
	}
}

func (s *Emotion) Init(resPath string, r *raspi.Adaptor) error {

	if err := s.term.Init(); err != nil {
		return err
	}
	// Initialize OLED displays.
	lock := &sync.Mutex{}
	if err := s.eye.Init(r, 1, 0x3c, lock, "eye"); err != nil {
		return err
	}
	if err := s.mouth.Init(r, 1, 0x3d, lock, "mouth"); err != nil {
		return err
	}

	// Start up "facial muscles".
	if err := s.term.Run(); err != nil {
		return err
	}
	if err := s.eye.Run(); err != nil {
		return err
	}
	if err := s.mouth.Run(); err != nil {
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
	s.Expression(EMOTION_NORM, CH, 1000)

	return nil
}

// CycleEmotions cycles through emotions; primarily a test function.
func (s *Emotion) CycleEmotions() {

	for k, _ := range s.faceEmotions {
		s.Expression(k, CH, 200)
		glog.V(2).Infof("Displaying emotion %v", k)
		time.Sleep(3000 * time.Millisecond)
	}
}

// Expression displays the requested emotion using the character ch. If the expression is
// animated it switchesusing ms milliseconds.
func (s *Emotion) Expression(emotion byte, ch rune, ms uint) error {

	e, ok := s.termEmotions[emotion]
	if !ok {
		return fmt.Errorf("expression not found")
	}
	s.term.Animate(e, CH, time.Duration(ms)*time.Millisecond)

	face, ok := s.faceEmotions[emotion]
	if !ok {
		return fmt.Errorf("expression not found")
	}
	s.eye.Animate(face.eye, ms)
	s.mouth.Animate(face.mouth, ms)

	return nil
}

func (s *Emotion) Quit() {
	s.term.Quit()
	s.eye.Quit()
	s.mouth.Quit()
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

	// Load Eye Expressions.
	eye, err := LoadImages(
		resPath + "/eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	eyeBlink, err := LoadImages(
		resPath+"/eye.png",
		resPath+"/eye_half_closed.png",
		resPath+"/eye_full_closed.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	eyeSide2Side, err := LoadImages(
		resPath+"/eye_look_left.png",
		resPath+"/eye.png",
		resPath+"/eye_look_right.png",
		resPath+"/eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	eyePupilDialated, err := LoadImages(
		resPath + "/wide_eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	eyePupilMov, err := LoadImages(
		resPath+"/eye.png",
		resPath+"/wide_eye.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	// Load mouth expressions.
	mouth, err := LoadImages(
		resPath + "/mouth.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthOpenSM, err := LoadImages(
		resPath + "/mouth_half_open.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthOpenLG, err := LoadImages(
		resPath + "/mouth_full_open.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthSpeak, err := LoadImages(
		resPath+"/mouth.png",
		resPath+"/mouth_half_open.png",
		resPath+"/mouth_full_open.png",
		resPath+"/mouth_half_open.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthSmileSM, err := LoadImages(
		resPath + "/mouth_half_smile.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthSmileLG, err := LoadImages(
		resPath + "/mouth_full_smile.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthInvSM, err := LoadImages(
		resPath + "/mouth_half_inverted.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	mouthInvLG, err := LoadImages(
		resPath + "/mouth_full_inverted.png",
	)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to load image resources: %v", err))
	}

	_ = eyeBlink

	return map[byte]Face{
		EMOTION_NORM:      Face{eye, mouth},
		EMOTION_SPEAK:     Face{eye, mouthSpeak},
		EMOTION_BLINK:     Face{eyePupilMov, mouthSmileSM},
		EMOTION_ANGRY:     Face{eye, mouthInvLG},
		EMOTION_SAD:       Face{eye, mouthInvSM},
		EMOTION_PUZZLED:   Face{eyePupilMov, mouthOpenLG},
		EMOTION_HAPPY:     Face{eyePupilDialated, mouthSmileLG},
		EMOTION_SMILE_MED: Face{eye, mouthSmileSM},
		EMOTION_THINKING:  Face{eyeSide2Side, mouthOpenSM},
	}, nil

}
