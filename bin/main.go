package main

import (
	"flag"

	"github.com/deepakkamesh/walle"
	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
)

func main() {

	secretsFile := flag.String("secrets_file", "/Users/dkg/Downloads/walle_prototype.json", "Path to secrets file")
	assistantScope := flag.String("assistant_scope", "https://www.googleapis.com/auth/assistant-sdk-prototype", "comma seperated list of scope urls for assistant")
	resourcesPath := flag.String("resources_path", "../resources", "Path to resources folder")

	flag.Parse()

	// Initialize Audio.
	aud := audio.New()

	// Initialize Google Assistant.
	ai := assistant.New(aud, *secretsFile, *assistantScope)

	// Build config for Walle.
	config := &walle.WallEConfig{
		Audio:      aud,
		GAssistant: ai,
	}

	wallE := walle.New(config)
	wallE.Run()

	glog.Infof("WallE Terminated")
}
