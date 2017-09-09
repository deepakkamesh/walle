package main

import (
	"flag"
	"time"

	"github.com/deepakkamesh/walle"
	"github.com/golang/glog"
)

func main() {

	secretsFile := flag.String("secrets_file", "walle_prototype.json", "Secrets file name in resources folder")
	assistantScope := flag.String("assistant_scope", "https://www.googleapis.com/auth/assistant-sdk-prototype", "comma seperated list of scope urls for assistant")
	resourcesPath := flag.String("resources_path", "../resources", "Path to resources folder")
	btnPort := flag.String("button_pin", "40", "Pin number for push button")
	irPort := flag.String("ir_pin", "38", "Pin number for IR")

	flag.Parse()

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
		AssistantScope: *assistantScope,
		SecretsFile:    *secretsFile,
		ResourcePath:   *resourcesPath,
		BtnPort:        *btnPort,
		IRPort:         *irPort,
	}

	wallE := walle.New()

	if err := wallE.Init(config); err != nil {
		glog.Fatalf("WallE initialization failed %v", err)
	}

	wallE.Run()

	// Needed so termbox can cleanup.
	t := time.NewTimer(10 * time.Millisecond)
	<-t.C

	glog.Infof("WallE esta muerto")
	glog.Flush()
}
