package main

import (
	"flag"
	"time"

	"github.com/deepakkamesh/termdraw"
	"github.com/deepakkamesh/walle"
	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
)

func main() {

	secretsFile := flag.String("secrets_file", "/Users/dkg/Downloads/walle_prototype.json", "Path to secrets file")
	assistantScope := flag.String("assistant_scope", "https://www.googleapis.com/auth/assistant-sdk-prototype", "comma seperated list of scope urls for assistant")
	resourcesPath := flag.String("resources_path", "../resources", "Path to resources folder")
	enEmotion := flag.Bool("en_emotion", true, "Enables WallE facial expression mode")

	flag.Parse()

	// Initialize Audio.
	aud := audio.New()

	// Initialize Google Assistant.
	ai := assistant.New(aud, *secretsFile, *assistantScope)

	// Enable WallE Face.
	var td *termdraw.Term
	if *enEmotion {
		td = termdraw.New()
	}

	// Build config for Walle.
	config := &walle.WallEConfig{
		Audio:        aud,
		GAssistant:   ai,
		Term:         td,
		ResourcePath: *resourcesPath,
	}

	wallE := walle.New(config)

	if err := wallE.Init(); err != nil {
		glog.Fatalf("Failure %v", err)
	}

	wallE.Run()

	// Needed so termbox can cleanup.
	t := time.NewTimer(10 * time.Millisecond)
	<-t.C
	glog.Infof("WallE Terminated")
	glog.Flush()
}
