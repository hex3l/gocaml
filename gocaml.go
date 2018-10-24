package main

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		Ocaml_StdinPipe(arg)
	} else {
		fmt.Fprintln(os.Stdout, "Missing .ml file! Usage: gocaml <ocaml file>")
	}
}

func Ocaml_StdinPipe(file string) {
	cmd := exec.Command("ocaml", "-init", file)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if nil != err {
		log.Fatalf("Error obtaining stdin: %s", err.Error())
	}
	cmd.Stdout = os.Stdout
	if nil != err {
		log.Fatalf("Error obtaining stdout: %s", err.Error())
	}
	writer := bufio.NewReader(os.Stdin)
	go FileNewWatcher(stdin, file)
	go func(writer io.Reader) {
		scanner := bufio.NewScanner(writer)
		for scanner.Scan() {
			line := scanner.Text()
			stdin.Write([]byte(line + "\n"))
		}
	}(writer)
	if err := cmd.Start(); nil != err {
		log.Fatalf("Error starting program: %s, %s", cmd.Path, err.Error())
	}
	cmd.Wait()
}

func FileNewWatcher(stdin io.WriteCloser, file string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Remove == fsnotify.Remove {
					err = watcher.Add(file)
					if err != nil {
						log.Println(err)
					}
					UpdateFile(stdin, file)
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					UpdateFile(stdin, file)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(file)
	if err != nil {
		log.Println(err)
	}
	<-done
}

func UpdateFile(stdin io.WriteCloser, file string) {
	os.Stdout.Write([]byte("\n"))
	stdin.Write([]byte("\n"))
	stdin.Write([]byte("#use \"" + file + "\";;" + "\n"))
}
