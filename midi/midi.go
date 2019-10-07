package midi

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
)

const programChangeStatusNum = uint8(0xC0)
const controlChangeStatusNum = uint8(0xB0)
const volumeControllerNum = uint8(0x07)

// NonEmphasizedTrackVolume the volume to set the non-emphasized tracks to
var NonEmphasizedTrackVolume = uint8(40)

// EmphasizedInstrumentNum the number corresponding to the instrument played by the emphasized track
var EmphasizedInstrumentNum = uint8(65)

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

	fmt.Println("total number of tracks", midi.GetTracksNum())
	// iterating through all tracks in MIDI file
	for currentTrackNum := uint16(0); currentTrackNum < midi.GetTracksNum(); currentTrackNum++ {
		curTrack := midi.GetTrack(currentTrackNum)
		isHeader, trackChannel := isHeaderTrackAndGetTrackChannel(curTrack)

		// if there is no MIDI_EVENT in the track, there's nothing to change in this track
		if isHeader {
			// we're not doing anything with this track, but we still want it to be included in the list
			fmt.Println("Header channel", trackChannel)
			tracksAtFullVolume = append(tracksAtFullVolume, curTrack)
			tracksWithLoweredVolume = append(tracksWithLoweredVolume, curTrack)
			continue
		}
		var eventPos = uint32(0)
		newInsIter := curTrack.GetIterator()
		newInstrumentEvent := createNewInstrumentEvent(curTrack, EmphasizedInstrumentNum, trackChannel)

		// replace the instrument on all the full-volume tracks
		for newInsIter.MoveNext() {
			if newInsIter.GetValue().GetStatus() == programChangeStatusNum {
				newInstrumentTrack := createNewTrack(curTrack, eventPos, newInstrumentEvent)
				tracksAtFullVolume = append(tracksAtFullVolume, newInstrumentTrack)
				break
			}
			eventPos++
		}

		// I would have thought that for channel 3, the controlChangeStatus number should be 0xB3, but
		// it seems the MIDIs created for ACC use 0xB0, 0x80, and 0x90 for all channels
		// controlChangeStatus := controlChangeStatusNum // + trackChannel

		// get all midi events via iterator
		iter := curTrack.GetIterator()
		volumeMIDIEvent := createNewVolumeEvent(curTrack, NonEmphasizedTrackVolume, trackChannel)

		eventPos = uint32(0)
		for iter.MoveNext() {
			// grab the track name so we can name our output files correctly
			if iter.GetValue().GetMetaType() == smf.MetaSequenceTrackName {
				trackNameMap[currentTrackNum] = grabTrackName(iter.GetValue())
			}
			// once we've found the MIDI event that's setting the channel volume, replace the old MIDI event with one that has the desired channel volume
			if iter.GetValue().GetStatus() == controlChangeStatusNum && iter.GetValue().GetData()[0] == volumeControllerNum {
				newVolumeTrack := createNewTrack(curTrack, eventPos, volumeMIDIEvent)
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

// Creates the output .mid files
func writeNewMIDIFile(wg *sync.WaitGroup, fileNum int, newMidiFile *smf.MIDIFile, trackNameMap map[uint16]string, midiFileName string) {
	defer wg.Done()
	var newFileName string
	// if the track didn't have a name (e.g., a track consisting only of META_EVENT's), we skip the .mid file creation
	if trackName, ok := trackNameMap[uint16(fileNum)]; ok {
		newFileName = "./output/" + midiFileName + "_" + trackName + ".mid"
	} else {
		return
	}

	fmt.Println("Creating", newFileName, "with all other tracks set to volume", NonEmphasizedTrackVolume)

	newpath := filepath.Join(".", "output")
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
