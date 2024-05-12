package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/radovskyb/watcher"
)

type SaveOutput struct {
}

func (so *SaveOutput) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(p)
}

var stdin io.Writer
var running bool = false
var reloading bool = false

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

	for {
		select {
		case event := <-w.Event:
			if running && !reloading && !event.IsDir() {
				fmt.Println(event.Path)
				go reload()
			}
		case err := <-w.Error:
			log.Fatalln(err)
		}
	}
}

func reload() {
	if reloading {
		return
	}
	reloading = true
	time.Sleep(time.Millisecond * 500)

	n, err := stdin.Write([]byte("r"))
	if err != nil {
		fmt.Println(err, n)
	}

	reloading = false
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
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				break
			}
			// Write the line to the io writer.
			stdin.Write([]byte(line))
		}
	}()
	fmt.Println("Flutter app is Running...")
	out.Wait()
	os.Exit(1)
}
