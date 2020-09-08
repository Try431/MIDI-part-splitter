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

// enabling line numbers in logging
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	binaryName := os.Args[0]

	fileFlagPtr := flag.String("f", "", "Name of .mid file you wish to parse\n(e.g., '"+binaryName+" -f midi_file.mid')")
	dirFlagPtr := flag.String("d", "", "Directory containing .mid files you wish to parse - will recursively search subdirectories\n(e.g., '"+binaryName+" -d ./dir/to/search/')")
	instFlagPtr := flag.Int("inst", 65, "Instrument number for emphasized track - see README for instrument list\n(e.g., '"+binaryName+" -f midi_file.mid -inst 22) ")
	volFlagPtr := flag.Int("vol", 40, "Volume of de-emphasized voice tracks - must be between 0 and 100\n(e.g., '"+binaryName+" -f midi_file.mid -vol 30)")
	outFlagPtr := flag.String("o", "./"+midi.MIDIOutputDirectory+"/mp3s", "Directory where mp3 files will be stored\n(e.g., '"+binaryName+" -f midi_file.mid -o ./dir/to/store/mp3s)")
	quietFlagPtr := flag.Bool("quiet", true, "Whether or not to silence standard output when running (will still allow stderr)")
	// listFlagPtr := flag.String("l", "", "List of comma-separated files")

	flag.Parse()

	if !isFlagPassed("f") && !isFlagPassed("d") {
		flag.Usage()
		os.Exit(1)
	}

	if isFlagPassed("quiet") {
		midi.SilenceOutput = bool(*quietFlagPtr)
	}

	// if isFlagPassed("l") {
	// 	// files := strings.Split(*listFlagPtr, ",")
	// 	// for _, f := range files {
	// 	// TODO - create helper function collectFilePaths that does what is done in the isFlagPassed("f") func, and use here and in that func to reduce code reuse
	// 	// Also, make sure that calling without a ./ or ../ prefix still works, as well as absolute filepaths
	// 	// }
	// }

	if isFlagPassed("vol") {
		midi.NonEmphasizedTrackVolume = uint8(*volFlagPtr)
	}

	if isFlagPassed("o") {
		midi.MP3OutputDirectory = *outFlagPtr
	}
	var filePaths []string

	if isFlagPassed("inst") {
		midi.EmphasizedInstrumentNum = uint8(*instFlagPtr)
	}

	if isFlagPassed("f") {
		filePath := filepath.Clean(*fileFlagPtr)
		extension := filepath.Ext(filePath)
		if !(strings.ToLower(extension) == ".mid" || strings.ToLower(extension) == ".midi") {
			log.Panicf("Extension " + extension + " not accepted - only MIDI files allowed")
		}
		filePaths = append(filePaths, filePath)
	}

	if isFlagPassed("d") {
		files := grabFilesInDir(*dirFlagPtr)
		filePaths = append(filePaths, files...)
	}

	fmt.Println("Starting split & conversion process...")
	var wg sync.WaitGroup
	wg.Add(len(filePaths))
	for i := 0; i < len(filePaths); i++ {
		fPath := filePaths[i]
		midi.SplitParts(&wg, fPath)
	}
	wg.Wait()
	fmt.Println("All done! 😄 Enjoy your MP3 files!")
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

// Walks through directory recursively and grabs all MIDI files
func grabFilesInDir(dirPath string) []string {
	var files []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && (filepath.Ext(info.Name()) == ".mid" || filepath.Ext(info.Name()) == ".midi") {
			filePath := filepath.Clean(path)
			files = append(files, filePath)
		}
		return nil
	})
	if err != nil {
		log.Panicf("Failed to walk through dirPath %v with error: %v", dirPath, err)
	}
	return files
}
