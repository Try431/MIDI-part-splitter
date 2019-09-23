package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/Try431/MIDI-part-splitter/midi"
)

const controlChangeStatusNum = uint8(0xB0)
const volumeControllerNum = uint8(0x07)

// const assetRoute = "./assets/"

var nonEmphasizedTrackVolume = uint8(40)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	binaryName := os.Args[0]

	fileFlagPtr := flag.String("f", "", "Name of .mid file you wish to parse\n(e.g., '"+binaryName+" -f midi_file.mid')")
	dirFlagPtr := flag.String("d", "", "Directory containing .mid files you wish to parse - will recursively search subdirectories\n(e.g., '"+binaryName+" -d ./dir/to/search/')")

	flag.Parse()

	var filePaths []string
	var fileNames []string

	if *fileFlagPtr == "" && *dirFlagPtr == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *fileFlagPtr != "" {
		dotSplit := strings.Split(*fileFlagPtr, ".")

		if len(dotSplit) == 1 {
			fmt.Println("Please supply name of file - do not forget file extension")
			os.Exit(1)
		}

		if strings.ToLower(dotSplit[len(dotSplit)-1]) != "mid" && strings.ToLower(dotSplit[len(dotSplit)-1]) != "midi" {
			fmt.Println("Only .mid and .midi files supported")
			os.Exit(1)
		}

		if !strings.HasPrefix(*fileFlagPtr, "./") {
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
			midiFilePath := dotSplit[1]
			filePaths = append(filePaths, midiFilePath)
		}
	}

	if *dirFlagPtr != "" {
		files := grabFilesInDir(*dirFlagPtr)
		filePaths = append(filePaths, files...)
	}
	fileNames = extractFileNamesFromPaths(filePaths)
	fmt.Println(fileNames)
	log.Fatal()

	midiFilePath := strings.Split(*fileFlagPtr, ".")[0]
	midiFileName := midiFilePath
	if strings.Contains(midiFilePath, "/") {
		split := strings.Split(midiFilePath, "/")
		midiFileName = split[len(split)-1]
	}

	// var wg sync.WaitGroup
	// wg.Add(len(filePaths))
	// for num, mFile := range filePaths {
	// 	go midi.SplitParts(&wg)
	// }
	// wg.Wait()

	midi.SplitParts(midiFilePath, midiFileName)
}

func grabFilesInDir(dirPath string) []string {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		midiFilePath := strings.Split(path, ".")[0]
		files = append(files, midiFilePath)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

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
