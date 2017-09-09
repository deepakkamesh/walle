package walle

import (
	"fmt"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/drivers/gpio"
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
	SecretsFile    string
	AssistantScope string
	ResourcePath   string
	BtnPort        string
	IRPort         string
}

type WallE struct {
	audio      *audio.Audio
	gAssistant *assistant.GAssistant // Google Assistant object.
	emotion    *Emotion
	btnChan    chan *gobot.Event
	irChan     chan *gobot.Event
	resPath    string
}

// New returns a new initialized WallE object.
func New() *WallE {

	return &WallE{
		audio:      audio.New(),
		gAssistant: assistant.New(),
		emotion:    NewEmotion(),
	}
}

// Init initializes WallE subsystems (gAssistant, Audio, ).
func (s *WallE) Init(c *WallEConfig) error {

	s.resPath = c.ResourcePath
	// Initialize Audio.
	if err := s.audio.Init(); err != nil {
		return err
	}
	s.audio.StartPlayback()

	// Initialize Google Assistant.
	if err := s.gAssistant.Init(s.audio, fmt.Sprintf("%v/%v", c.ResourcePath, c.SecretsFile), c.AssistantScope); err != nil {
		return err
	}
	if err := s.gAssistant.Auth(); err != nil {
		return err
	}

	// Initialize Pi Adapter.
	rpi := raspi.NewAdaptor()
	if err := rpi.Connect(); err != nil {
		return err
	}

	// Init Emotion controller.
	if err := s.emotion.Init(c.ResourcePath, rpi); err != nil {
		return fmt.Errorf("failed to init emotions:%v", err)
	}

	// Initialize pushbutton.
	button := gpio.NewButtonDriver(rpi, c.BtnPort)
	if err := button.Start(); err != nil {
		return err
	}
	s.btnChan = button.Subscribe()

	// Initialize IR Sensor.
	ir := gpio.NewButtonDriver(rpi, c.IRPort)
	if err := ir.Start(); err != nil {
		return err
	}
	s.irChan = ir.Subscribe()

	return nil
}

// Run is the main event loop.
func (s *WallE) Run() {
	for {
		select {
		// Events from termui for keyboard events.
		case evt := <-s.emotion.term.EventCh:
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

				case evt.Ch == 's':
					TextToSpeech(s.resPath+"/bored.raw", s.audio)
				}
			}

		case evt := <-s.btnChan:
			glog.V(2).Infof("Got event from pushbutton %v-%v", evt.Name, evt.Data)
			if evt.Name == "push" {
				s.interactAI()
			}

		case evt := <-s.irChan:
			glog.V(2).Infof("Got event from IR proximity sensor %v-%v", evt.Name, evt.Data)
		}
	}
	return
}

// interactAI runs a gAssistant session collects the response text
// and analyzes it for sentiment.
func (s *WallE) interactAI() {

	//TODO: This is a workaround for Pi as the audio does not continue playing after
	// first interaction. Needs investigation and fix.
	s.audio.ResetPlayback()

	go func() {
		st := <-s.gAssistant.StatusCh
		if st == embedded.ConverseResponse_END_OF_UTTERANCE {
			glog.V(2).Infof("gAssisant sent END_OF_UTTERANCE")
			s.emotion.Expression(EMOTION_SPEAK, CH, 200)
		}
	}()

	s.emotion.Expression(EMOTION_BLINK, CH, 100)
	audioOut := s.gAssistant.ConverseWithAssistant()
	s.emotion.Expression(EMOTION_THINKING, CH, 100)

	// Convert assistant audio to text.
	txt, err := SpeechToText(audioOut)
	if err != nil {
		glog.Errorf("Failed to recognize speech: %v", err)
		s.emotion.Expression(EMOTION_SAD, CH, 900)
		return
	}
	glog.V(1).Infof("Google Assistant said: %v", txt)

	// Get sentiment analysis of text.
	score, magnitude, err := AnalyzeSentiment(txt)
	if err != nil {
		glog.Errorf("Failed to analyze sentiment: %v", err)
		s.emotion.Expression(EMOTION_SAD, CH, 900)
		return
	}
	glog.V(1).Infof("Sentiment Analysis - Score:%v Magnitude:%v", score, magnitude)

	// Select an emotion to display.
	emotion := selectEmotion(score, txt)
	s.emotion.Expression(emotion, CH, 300)
}
