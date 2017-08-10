package main

import (
	"flag"
	"walle"
	"walle/assistant"
	"walle/audio"

	"github.com/golang/glog"
)

func main() {

	secretsFile := flag.String("secrets_file", "/Users/dkg/Downloads/walle_prototype.json", "Path to secrets file")
	assistantScope := flag.String("assistant_scope", "https://www.googleapis.com/auth/assistant-sdk-prototype", "comma seperated list of scope urls for assistant")
	flag.Parse()

	// Initialize Audio.
	aud := audio.New()

	// Initialize Google Assistant.
	ai := assistant.New(aud.Out, aud.In, *secretsFile, *assistantScope)

	// Build config for Walle.
	config := &walle.WallEConfig{
		Audio:      aud,
		GAssistant: ai,
	}

	wallE := walle.New(config)
	wallE.Run()

	glog.Infof("WallE Terminated")
}
