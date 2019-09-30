package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Try431/MIDI-part-splitter/midi"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	binaryName := os.Args[0]

	fileFlagPtr := flag.String("f", "", "Name of .mid file you wish to parse\n(e.g., '"+binaryName+" -f midi_file.mid')")
	dirFlagPtr := flag.String("d", "", "Directory containing .mid files you wish to parse - will recursively search subdirectories\n(e.g., '"+binaryName+" -d ./dir/to/search/')")
	volFlagPtr := flag.Int("vol", 40, "[Optional] Volume of de-emphasized voice tracks - must be between 0 and 100")

	flag.Parse()

	if isFlagPassed("vol") {
		midi.NonEmphasizedTrackVolume = uint8(*volFlagPtr)
	}
	var filePaths []string
	var fileNames []string
	var extensions []string

	if !isFlagPassed("f") && !isFlagPassed("d") {
		flag.Usage()
		os.Exit(1)
	}

	if isFlagPassed("f") {
		dotSplit := strings.Split(*fileFlagPtr, ".")

		if len(dotSplit) == 1 {
			fmt.Println("Please supply name of file - do not forget file extension")
			os.Exit(1)
		}

		if strings.ToLower(dotSplit[len(dotSplit)-1]) != "mid" && strings.ToLower(dotSplit[len(dotSplit)-1]) != "midi" {
			fmt.Println("Only .mid and .midi files supported")
			os.Exit(1)
		} else {
			extensions = append(extensions, strings.ToLower(dotSplit[len(dotSplit)-1]))
		}

		if strings.HasPrefix(*fileFlagPtr, "../") {
			midiFilePath := strings.Split(*fileFlagPtr, ".mid")[0]
			filePaths = append(filePaths, midiFilePath)
		} else if !strings.HasPrefix(*fileFlagPtr, "./") {
			if len(dotSplit) != 2 {
				fmt.Println("Filename has more than one \".\" - please fix")
				os.Exit(1)
			}
			midiFilePath := dotSplit[0]
			filePaths = append(filePaths, midiFilePath)

		} else if strings.HasPrefix(*fileFlagPtr, "./") {
			if len(dotSplit) != 3 {
				fmt.Println("Filename has more than one \".\" - please fix")
				os.Exit(1)
			}
			midiFilePath := "." + dotSplit[1]
			filePaths = append(filePaths, midiFilePath)
		}
	}

	if isFlagPassed("d") {
		files, exts := grabFilesInDir(*dirFlagPtr)
		filePaths = append(filePaths, files...)
		extensions = append(extensions, exts...)
	}
	fileNames = extractFileNamesFromPaths(filePaths)

	if len(fileNames) != len(filePaths) && len(filePaths) != len(extensions) {
		log.Panicf("Mismatched number of file names, file paths, and file extensions")
	}
	var wg sync.WaitGroup
	wg.Add(len(filePaths))
	for i := 0; i < len(fileNames); i++ {
		fPath := filePaths[i]
		fName := fileNames[i]
		ext := extensions[i]
		go midi.SplitParts(&wg, fPath, fName, ext)
	}
	wg.Wait()
}

// Determines if a flag was passed in
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// Walks through directory recursively and grabs all .mid files
func grabFilesInDir(dirPath string) ([]string, []string) {
	var files []string
	var exts []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		midiFilePath := strings.Split(path, ".mid")[0]
		dotSplit := strings.Split(path, ".")
		extension := dotSplit[len(dotSplit)-1]
		files = append(files, midiFilePath)
		exts = append(exts, extension)
		return nil
	})
	if err != nil {
		log.Panicf("Failed to walk through dirPath %v with error: %v", dirPath, err)
	}
	return files, exts
}

// Grabs the actual .mid or .midi filename from a filepath
func extractFileNamesFromPaths(filePaths []string) []string {
	var fileNames []string
	for _, path := range filePaths {
		if strings.Contains(path, "/") {
			split := strings.Split(path, "/")
			fileNames = append(fileNames, split[len(split)-1])
		} else {
			fileNames = append(fileNames, path)
		}
	}
	return fileNames
}
