package walle

import (
	"bytes"
	"context"
	"io/ioutil"

	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
)

const (
	CHUNK_SZ = 1600
)

// TextToSpeech plays the raw audio file identified by fname.
func TextToSpeech(fname string, aud *audio.Audio) error {
	aud.ResetPlayback()
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return err
	}
	l := len(data)
	for i := 0; i < l; i += CHUNK_SZ {
		end := i + CHUNK_SZ
		if end > l {
			end = l
		}
		chunk := data[i:end]
		buf := bytes.NewBuffer(chunk)
		aud.Out <- *buf
	}
	return nil
}

func SpeechToText(audio *bytes.Buffer) (resultTxt string, err error) {
	ctx := context.Background()

	// Creates a client.
	client, err := speech.NewClient(ctx)
	if err != nil {
		return
	}

	// Detects speech in the audio data.
	req := &speechpb.RecognizeRequest{
		Config: &speechpb.RecognitionConfig{
			Encoding:        speechpb.RecognitionConfig_LINEAR16,
			SampleRateHertz: 16000,
			LanguageCode:    "en-US",
		},
		Audio: &speechpb.RecognitionAudio{
			AudioSource: &speechpb.RecognitionAudio_Content{Content: audio.Bytes()},
		},
	}
	resp, err := client.Recognize(ctx, req)
	if err != nil {
		return
	}

	// Prints the results.
	for _, result := range resp.Results {
		for _, alt := range result.Alternatives {
			resultTxt = alt.Transcript
			glog.V(3).Infof("\"%v\" (confidence=%3f)\n", alt.Transcript, alt.Confidence)
		}
	}

	return

}
