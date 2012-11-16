package main

import (
	"log"
	"os/exec"
	"os"
	"path/filepath"
	"time"
)

var (
	command string
	directory string
	modification_time time.Time
	access_time time.Time
	highest_modification int64
	cmd *exec.Cmd
)

func main() {
	if len(os.Args) < 2 {
		log.Printf("\033[1;31mPlease provide application (usage: gomon <path_to_application>)\033[0m\n")
		return
	} 

	command = os.Args[1]
	directory = filepath.Dir(command)
	check_files_for_changes()

	log.Printf("\033[37mMonitoring: %s\033[0m\n", directory)
	c := time.Tick(1 * time.Second)

	cmd = exec.Command(command)
	cmd.Dir = directory
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	for now := range c {
		if check_files_for_changes() && now.Unix() > 0 {

			log.Printf("\033[1;32mRebuilding application!\033[0m\n")

			err := cmd.Process.Kill()
			if err != nil {
				log.Printf("\033[1;31mError killing running process: %s\033[0m\n", err)
			}

			rebuild_cmd := exec.Command("go", "build")
			rebuild_cmd.Dir = directory
			command_output, err_out := rebuild_cmd.Output()

			errn := rebuild_cmd.Run()
			if errn != nil && errn.Error() != "exec: already started" {
				log.Printf("\033[1;31mError rebuilding application: %s\033[0m\n", errn)
			}
			if err_out != nil {
				log.Printf("\033[1;31m=== Ouput from application ====\n\n%s\n", command_output)
				log.Printf("=== End of ouput ====\033[0m\n")
			}

			cmd = nil
			cmd = exec.Command(command)
			cmd.Dir = directory

			err = cmd.Start()
			if err != nil {
				log.Printf("\033[1;31mError starting application: %s\033[0m\n", err)
			}

		}
	}
}

func check_files_for_changes() bool {
	is_modified := false
	
	checkFunc := func(path string, info os.FileInfo, err error) error {
		ex := filepath.Ext(path)
		if ex == ".go" {
			f, e := os.Open(path)
			if e != nil {
				log.Printf("\033[1;31mError accessing file: %s\033[0m\n", e)
			}

			stats, e2 := f.Stat()
			if e2 != nil {
				log.Printf("\033[1;31mError getting stats from opened file: %s\033[0m\n", e2)
			}

			unix_timestamp := stats.ModTime().Unix()

			if unix_timestamp > highest_modification {
				highest_modification = unix_timestamp
				is_modified = true
			}	
		}
		return err
	}

	errn := filepath.Walk(directory, checkFunc)
	if errn != nil {
		log.Printf("\033[1;31mError walking directory: %s\033[0m\n", errn)
	}

	return is_modified
}