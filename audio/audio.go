package audio

import (
	"bytes"
	"encoding/binary"

	"github.com/golang/glog"
	"github.com/gordonklaus/portaudio"
)

type Audio struct {
	In        chan bytes.Buffer
	Out       chan bytes.Buffer
	streamIn  *portaudio.Stream
	streamOut *portaudio.Stream
	bufIn     []int16
	bufOut    []int16
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
		In:        make(chan bytes.Buffer),
		Out:       make(chan bytes.Buffer, 10),
		streamIn:  in,
		streamOut: out,
		bufIn:     bufIn,
		bufOut:    bufOut,
	}
}

func (s *Audio) Speak(text []byte) {

}

func (s *Audio) Run() {
	go s.listen()
	go s.playback()
}

func (s *Audio) listen() {
	if err := s.streamIn.Start(); err != nil {
		glog.Fatalf("Failed to connect to the input stream", err)
	}

	for {
		if err := s.streamIn.Read(); err != nil {
			glog.Errorf("Failed to read input stream: %v", err)
		}

		var bufWriter bytes.Buffer
		binary.Write(&bufWriter, binary.LittleEndian, s.bufIn)
		s.In <- bufWriter
	}
}

func (s *Audio) playback() {
	if err := s.streamOut.Start(); err != nil {
		glog.Fatalf("Failed to start audio out")
	}

	for {
		out := <-s.Out

		if err := binary.Read(&out, binary.LittleEndian, s.bufOut); err != nil {
			glog.Errorf("Failed to convert to binary %v", err)
		}

		glog.V(3).Infof("Audio chunk size: %v", len(s.bufOut))
		if err := s.streamOut.Write(); err != nil {
			glog.Warningf("Failed to write to audio out: %v", err)
		}
	}
}

func (s *Audio) StopListening() error {
	return s.streamIn.Stop()
}

func (s *Audio) Quit() {
	s.streamIn.Close()
	s.streamOut.Close()
}
