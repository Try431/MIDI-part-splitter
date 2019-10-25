package midi

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
)

const programChangeStatusNum = uint8(0xC0)
const controlChangeStatusNum = uint8(0xB0)
const volumeControllerNum = uint8(0x07)

// MIDIOutputDirectory the directory where the converted MIDI files will be stored
var MIDIOutputDirectory = "output"

// NonEmphasizedTrackVolume the volume to set the non-emphasized tracks to
var NonEmphasizedTrackVolume = uint8(40)

// EmphasizedTrackVolume we must set the emphasized track volume to 100 because some MIDI tracks have non-100 default volumes
const EmphasizedTrackVolume = uint8(100)

// EmphasizedInstrumentNum the number corresponding to the instrument played by the emphasized track
var EmphasizedInstrumentNum = uint8(65)

// MP3OutputDirectory the directory where the mp3 files will be stored
var MP3OutputDirectory = "output/mp3s"

// outputMIDIFilePaths is a slice of the full filepaths of the MIDI files created by writeNewMIDIFile() -- this slice will be accessed by the conversion bash script
var outputMIDIFilePaths []string
var filepathLock sync.RWMutex

// SplitParts splits the MIDI file into different voice parts and creates new MIDI files
// with those voice parts emphasized
func SplitParts(mainWg *sync.WaitGroup, midiFilePath string, midiFileName string, extension string) {
	defer mainWg.Done()
	fullFilePath := midiFilePath + "." + extension
	file, err := os.Open(midiFilePath + "." + extension)
	if err != nil {
		log.Fatalf("Failed to open %v with error: %v", fullFilePath, err)
	}
	defer file.Close()

	// read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))
	if err != nil {
		log.Panicf("Failed to read MIDI file %v with error: %v", file, err)
	}

	// collecting record of all tracks in the MIDI file so we can construct our new MIDI files in the same track order
	var tracksWithLoweredVolume []*smf.Track
	var tracksAtFullVolume []*smf.Track
	trackNameMap := make(map[uint16]string)

	// iterating through all tracks in MIDI file
	for currentTrackNum := uint16(0); currentTrackNum < midi.GetTracksNum(); currentTrackNum++ {
		curTrack := midi.GetTrack(currentTrackNum)
		isHeader, trackChannel := isHeaderTrackAndGetTrackChannel(curTrack)

		// if there is no MIDI_EVENT in the track, there's nothing to change in this track
		if isHeader {
			// we're not doing anything with this track, but we still want it to be included in the list
			tracksAtFullVolume = append(tracksAtFullVolume, curTrack)
			tracksWithLoweredVolume = append(tracksWithLoweredVolume, curTrack)
			continue
		}
		var eventPos = uint32(0)
		newInsIter := curTrack.GetIterator()
		newInstrumentEvent := createNewInstrumentEvent(curTrack, EmphasizedInstrumentNum, trackChannel)
		var newInstrumentTrack *smf.Track
		// replace the instrument on all the full-volume tracks
		for newInsIter.MoveNext() {
			if newInsIter.GetValue().GetStatus() == programChangeStatusNum {
				newInstrumentTrack = createNewTrack(curTrack, eventPos, newInstrumentEvent)
				// newVolAndInstrumentTrack := createNewTrack(newInstrumentTrack, eventPos, highVolumeEvent)
				// tracksAtFullVolume = append(tracksAtFullVolume, newVolAndInstrumentTrack)
				break
			}
			eventPos++
		}

		eventPos = uint32(0)
		highVolIter := newInstrumentTrack.GetIterator()
		highVolumeEvent := createNewVolumeEvent(curTrack, EmphasizedTrackVolume, trackChannel)
		for highVolIter.MoveNext() {
			if highVolIter.GetValue().GetStatus() == controlChangeStatusNum && highVolIter.GetValue().GetData()[0] == volumeControllerNum {
				newVolAndInstrumentTrack := createNewTrack(newInstrumentTrack, eventPos, highVolumeEvent)
				tracksAtFullVolume = append(tracksAtFullVolume, newVolAndInstrumentTrack)
				break
			}
			eventPos++
		}

		// I would have thought that for channel 3, the controlChangeStatus number should be 0xB3, but
		// it seems the MIDIs created for ACC use 0xB0, 0x80, and 0x90 for all channels
		// controlChangeStatus := controlChangeStatusNum // + trackChannel

		// get all midi events via iterator
		iter := curTrack.GetIterator()
		lowVolumeMIDIEvent := createNewVolumeEvent(curTrack, NonEmphasizedTrackVolume, trackChannel)

		eventPos = uint32(0)
		for iter.MoveNext() {
			// grab the track name so we can name our output files correctly
			if iter.GetValue().GetMetaType() == smf.MetaSequenceTrackName {
				trackNameMap[currentTrackNum] = grabTrackName(iter.GetValue())
			}
			// once we've found the MIDI event that's setting the channel volume, replace the old MIDI event with one that has the desired channel volume
			if iter.GetValue().GetStatus() == controlChangeStatusNum && iter.GetValue().GetData()[0] == volumeControllerNum {
				newVolumeTrack := createNewTrack(curTrack, eventPos, lowVolumeMIDIEvent)
				tracksWithLoweredVolume = append(tracksWithLoweredVolume, newVolumeTrack)
				break
			}
			eventPos++
		}

	}

	var newMIDIFilesToBeCreated []*smf.MIDIFile
	var emphasizedTrackNum = uint16(0)
	for i := 0; i < len(tracksAtFullVolume); i++ {
		// create division
		division, err := smf.NewDivision(960, smf.NOSMTPE)
		if err != nil {
			log.Printf("Failed to create new Division object with error: %v", err)
		}

		// create new midi struct
		newMIDIFile, err := smf.NewSMF(smf.Format1, *division)
		if err != nil {
			log.Printf("Failed to create new MIDI object with error: %v", err)
		}

		fullVolTrack := tracksAtFullVolume[emphasizedTrackNum]
		for k := 0; k < len(tracksWithLoweredVolume); k++ {
			if uint16(k) == emphasizedTrackNum {
				newMIDIFile.AddTrack(fullVolTrack)
			} else {
				newMIDIFile.AddTrack(tracksWithLoweredVolume[k])
			}
		}
		newMIDIFilesToBeCreated = append(newMIDIFilesToBeCreated, newMIDIFile)
		emphasizedTrackNum++
	}

	var wg sync.WaitGroup
	wg.Add(len(newMIDIFilesToBeCreated))
	for num, mFile := range newMIDIFilesToBeCreated {
		go writeNewMIDIFile(&wg, num, mFile, trackNameMap, midiFileName)
	}
	wg.Wait()

	fmt.Println("\n%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
	fmt.Println("%%%%%% Beginning MIDI --> WAV --> MP3 conversion %%%%%%")
	fmt.Println("%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")

	var convertWg sync.WaitGroup
	convertWg.Add(len(outputMIDIFilePaths))
	for _, filepath := range outputMIDIFilePaths {
		go runConversionScript(&convertWg, filepath)
	}
	convertWg.Wait()
	fmt.Println("All done! 😄")
}

func runConversionScript(wg *sync.WaitGroup, filepath string) {
	defer wg.Done()
	cmd := exec.Command("/bin/sh", "convert/convert_async.sh", filepath, MP3OutputDirectory)
	// create a pipe for the output of the script
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating StdoutPipe for Cmd", err)
		return
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
		return
	}
}

// This function is primarily for debugging purposes, to check the volume of a track
func checkVolumeOfTrack(track *smf.Track) (uint8, int, map[int][]byte) {
	iter := track.GetIterator()
	var vol uint8
	var volCounts int
	volMap := make(map[int][]byte)
	pos := 0
	for iter.MoveNext() {
		if iter.GetValue().GetStatus() == controlChangeStatusNum && iter.GetValue().GetData()[0] == volumeControllerNum {
			volCounts++
			volMap[pos] = iter.GetValue().GetData()
			vol = uint8(iter.GetValue().GetData()[1])
		}
		pos++
	}
	return vol, volCounts, volMap
}

func handleDuplicateTrackNames(trackNameMap map[uint16]string) map[uint16]string {
	nameMap := make(map[string]int)
	newTrackNameMap := make(map[uint16]string)

	// we're creating a sorted slice of filenums here because iterating through trackNameMap directly
	// isn't guaranteed to be in the proper order - this solves the problem of incorrectly numbering a duplicate track name
	keys := make([]uint16, 0)
	for k := range trackNameMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		trackName := trackNameMap[k]
		nameMap[trackName]++
		if nameMap[trackName] > 1 {
			count := strconv.FormatInt(int64(nameMap[trackName]), 10)
			newTrackNameMap[k] = trackName + "(" + (count) + ")"
		} else {
			newTrackNameMap[k] = trackName
		}

	}

	return newTrackNameMap
}

// Creates the output .mid files
func writeNewMIDIFile(wg *sync.WaitGroup, fileNum int, newMidiFile *smf.MIDIFile, trackNameMap map[uint16]string, midiFileName string) {
	defer wg.Done()
	var newFileName string
	trackNameMap = handleDuplicateTrackNames(trackNameMap)
	// if the track didn't have a name (e.g., a track consisting only of META_EVENT's), we skip the .mid file creation
	if trackName, ok := trackNameMap[uint16(fileNum)]; ok {
		newFileName = "./" + MIDIOutputDirectory + "/" + midiFileName + "_" + trackName + ".mid"
	} else {
		return
	}

	fmt.Println("Creating", newFileName, "with all other tracks set to volume", NonEmphasizedTrackVolume)

	newpath := filepath.Join(".", MIDIOutputDirectory)
	err := os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Panicf("Failed to create directory %v with error: %v", newFileName, err)
	}
	outputMidi, err := os.Create(newFileName)
	if err != nil {
		log.Panicf("Failed to create new MIDI file with error: %v", err)
	}
	defer outputMidi.Close()

	writer := bufio.NewWriter(outputMidi)
	smfio.Write(writer, newMidiFile)
	writer.Flush()
	filepathLock.Lock()
	outputMIDIFilePaths = append(outputMIDIFilePaths, newFileName)
	filepathLock.Unlock()
}

// Parses hex bytes into text
func grabTrackName(e smf.Event) string {
	var trackName string
	for _, c := range e.GetData() {
		char := fmt.Sprintf("%c", c)
		trackName += char
	}
	// replace all spaces with underscores
	trackName = strings.ReplaceAll(trackName, " ", "_")
	return trackName
}

// Checks if there are any MIDI events in the track
func isHeaderTrackAndGetTrackChannel(track *smf.Track) (bool, uint8) {
	allEvents := track.GetAllEvents()
	headerEvent := true
	var chanNum uint8
	for _, e := range allEvents {
		if strings.HasPrefix(e.String(), "MIDI") {
			headerEvent = false
			chanNum = e.GetChannel()
			break
		}
	}
	return headerEvent, chanNum
}

// Returns a new MIDI smf.Track object with a specific event replaced
func createNewTrack(track *smf.Track, replacePos uint32, newEvent *smf.MIDIEvent) *smf.Track {
	allTrackEvents := track.GetAllEvents()
	allTrackEvents[replacePos] = newEvent
	var pos = uint32(0)
	var volCount = 0
	for _, event := range allTrackEvents {
		// if there's another volume control MIDI event in the track, we want to delete it, otherwise the changes we've made will be overridden
		if event.GetStatus() == controlChangeStatusNum && event.GetData()[0] == volumeControllerNum {
			volCount++
			if volCount > 1 {
				// delete track event at pos
				allTrackEvents = append(allTrackEvents[:pos], allTrackEvents[pos+1:]...)
			}
		}
		pos++
	}

	// create a new track with our updated array of events
	updatedTrack, err := smf.TrackFromArray(allTrackEvents)
	if err != nil {
		log.Panicf("Failed to create new track from event list with error: %v", err)
	}

	return updatedTrack
}

// Creates a new MIDI_EVENT to set the volume of the track we want to de-emphasize
func createNewVolumeEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatusNum, channel, volumeControllerNum, newVolume)
	if err != nil {
		log.Panicf("Failed to create new volume MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}

func createNewInstrumentEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	newInstrumentEvent, err := smf.NewMIDIEvent(0, programChangeStatusNum, channel, EmphasizedInstrumentNum, 0)
	if err != nil {
		log.Panicf("Failed to create new instrument MIDI event with error: %v", err)
	}
	return newInstrumentEvent
}
