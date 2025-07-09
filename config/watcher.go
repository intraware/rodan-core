package config

import (
	"fmt"
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

func WatchConfig(path string, reloadFunc func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}
	logrus.Println("Watching config for changes...")
	var debounceTimer *time.Timer
	debounceDelay := 5 * time.Second
	for {
		select {
		case event, ok := <-watcher.Events:
			fmt.Println(event, ok)
			if !ok {
				return
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					logrus.Println("Debounced config reload...")
					reloadFunc()
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logrus.Println("Watcher error:", err)
		}
	}
}
