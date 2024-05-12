package main

import (
	"fmt"
	"github.com/radovskyb/watcher"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type SaveOutput struct {
}

func (so *SaveOutput) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

var stdin io.Writer
var running bool = false

func main() {
	go flutterRun()

	w := watcher.New()
	go watchFiles(w)

	w.Start(time.Millisecond * 100)
}
func watchFiles(w *watcher.Watcher) {
	defer w.Close()
	w.FilterOps(watcher.Write, watcher.Create, watcher.Remove, watcher.Rename)
	w.AddRecursive("./lib")

	reloading := false

	for {
		select {
		case event := <-w.Event:
			if running == true && !reloading && !event.IsDir() {
				reloading = true
				time.Sleep(time.Millisecond * 500)

				n, err := stdin.Write([]byte("r"))
				if err != nil {
					fmt.Println(err, n)
				}

				reloading = false
			}
		case err := <-w.Error:
			log.Fatalln(err)
		}
	}
}

func flutterRun() {
	var so SaveOutput
	args := []string{"run"}
    args = append(args, os.Args[1:]...)

	out := exec.Command("flutter", args...)

	stdinPipe, err := out.StdinPipe()
	if err != nil {
		fmt.Println(err)
		return
	}
	stdin = stdinPipe

	out.Stdout = &so
	out.Stderr = os.Stderr

	startErr := out.Start()
	if startErr != nil {
		log.Fatalln(startErr)
		os.Exit(1)
	}

	running = true

	fmt.Println("Flutter app is Running...")
	outError := out.Wait()

	if outError != nil {
		os.Exit(1)
	}

}
