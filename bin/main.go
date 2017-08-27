package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/deepakkamesh/walle"
	"github.com/deepakkamesh/walle/assistant"
	"github.com/deepakkamesh/walle/audio"
	"github.com/golang/glog"
)

func main() {

	secretsFile := flag.String("secrets_file", "walle_prototype.json", "Secrets file name in resources folder")
	assistantScope := flag.String("assistant_scope", "https://www.googleapis.com/auth/assistant-sdk-prototype", "comma seperated list of scope urls for assistant")
	resourcesPath := flag.String("resources_path", "../resources", "Path to resources folder")

	flag.Parse()

	// Initialize Audio.
	aud := audio.New()

	// Initialize Google Assistant.
	ai := assistant.New(aud, fmt.Sprintf("%v/%v", *resourcesPath, *secretsFile), *assistantScope)

	// Flush logs to disk.
	logFlusher := time.NewTicker(300 * time.Millisecond)
	go func() {
		for {
			<-logFlusher.C
			glog.Flush()
		}
	}()

	// Build config for Walle.
	config := &walle.WallEConfig{
		Audio:        aud,
		GAssistant:   ai,
		ResourcePath: *resourcesPath,
	}

	wallE := walle.New(config)

	if err := wallE.Init(); err != nil {
		glog.Fatalf("WallE initialization failed %v", err)
	}

	wallE.Run()

	// Needed so termbox can cleanup.
	t := time.NewTimer(10 * time.Millisecond)
	<-t.C

	glog.Infof("WallE esta muerto")
	glog.Flush()
}
