package audio

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/golang/glog"
	"github.com/gordonklaus/portaudio"
)

type Audio struct {
	In           chan bytes.Buffer
	Out          chan bytes.Buffer
	streamIn     *portaudio.Stream
	streamOut    *portaudio.Stream
	bufIn        []int16
	bufOut       []int16
	listenStop   chan struct{}
	playbackStop chan struct{}
}

func New() *Audio {
	portaudio.Initialize()

	// Open Input stream.
	bufIn := make([]int16, 8196)
	in, err := portaudio.OpenDefaultStream(1, 0, 16000, len(bufIn), bufIn)
	if err != nil {
		glog.Fatalf("Failed to connect to the set the default stream", err)
	}

	// Open Output stream.
	bufOut := make([]int16, 799)
	out, err := portaudio.OpenDefaultStream(0, 1, 16000, len(bufOut), bufOut)
	if err != nil {
		glog.Fatalf("Failed to connect to the set the default stream", err)
	}

	return &Audio{
		In:           make(chan bytes.Buffer, 10),
		Out:          make(chan bytes.Buffer, 1000),
		streamIn:     in,
		streamOut:    out,
		bufIn:        bufIn,
		bufOut:       bufOut,
		listenStop:   make(chan struct{}),
		playbackStop: make(chan struct{}),
	}
}

func (s *Audio) Speak(text []byte) {
}

func (s *Audio) StartPlayback() {
	go s.playback()
}

func (s *Audio) StartListen() {
	go s.listen()
}

func (s *Audio) StopListen() {
	s.listenStop <- struct{}{}
}

func (s *Audio) StopPlayback() {
	s.playbackStop <- struct{}{}
}

// ResetPlayback resets the output stream (stop, start)
func (s *Audio) ResetPlayback() {
	s.StopPlayback()
	t := time.NewTimer(50 * time.Millisecond)
	<-t.C
	s.StartPlayback()
}

func (s *Audio) listen() {

	// TODO: Get a cleaner solution to by removing the buffered channel and
	// handling the race condition. Write to s.In in a go func.
	if len(s.In) > 0 {
		glog.Warningf("Audio input channel is non zero: %v", len(s.In))
	}
	if err := s.streamIn.Start(); err != nil {
		glog.Fatalf("Failed to start input stream: %v ", err)
	}

	listenFunc := func() {
		if err := s.streamIn.Read(); err != nil {
			glog.Errorf("Failed to read input stream: %v", err)
		}

		var bufWriter bytes.Buffer
		binary.Write(&bufWriter, binary.LittleEndian, s.bufIn)
		s.In <- bufWriter
	}

	for {
		select {
		case <-s.listenStop:
			if err := s.streamIn.Stop(); err != nil {
				glog.Errorf("Failed to stop input audio stream: %v", err)
			}
			return

		default:
			listenFunc()
		}
	}
}

func (s *Audio) playback() {
	if err := s.streamOut.Start(); err != nil {
		glog.Fatalf("Failed to start audio out: %v", err)
	}

	for {
		select {

		case <-s.playbackStop:
			if err := s.streamOut.Stop(); err != nil {
				glog.Errorf("Failed to stop output audio stream: %v", err)
			}
			return

		case out := <-s.Out:
			glog.V(3).Infof("Audio chunk size: %v", out.Len())
			if err := binary.Read(&out, binary.LittleEndian, s.bufOut); err != nil {
				glog.Warningf("Failed to convert to binary %v", err)
				continue
			}
			if err := s.streamOut.Write(); err != nil {
				glog.Warningf("Failed to write to audio out: %v", err)
			}

		}
	}
}

func (s *Audio) Quit() {
	if err := s.streamOut.Close(); err != nil {
		glog.Errorf("Failed to close output audio stream: %v", err)
	}
	if err := s.streamIn.Close(); err != nil {
		glog.Errorf("Failed to close input audio stream: %v", err)
	}
}
