package midi

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
)

const controlChangeStatusNum = uint8(0xB0)
const volumeControllerNum = uint8(0x07)

var nonEmphasizedTrackVolume = uint8(40)

func SplitParts(midiFilePath string, midiFileName string) {
	file, _ := os.Open(midiFilePath + ".mid")
	defer file.Close()

	// Read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))
	if err != nil {
		log.Panicf("Failed to read MIDI file %v with error: %v", file, err)
	}

	// Collecting record of all tracks in the MIDI file so we can construct our new MIDI files in the same track order
	var tracksWithLoweredVolume []*smf.Track
	var tracksAtFullVolume []*smf.Track
	trackNameMap := make(map[uint16]string)

	// iterating through all tracks in MIDI file
	for currentTrackNum := uint16(0); currentTrackNum < midi.GetTracksNum(); currentTrackNum++ {
		curTrack := midi.GetTrack(currentTrackNum)
		tracksAtFullVolume = append(tracksAtFullVolume, curTrack)
		// if there is no MIDI_EVENT in the track, there's nothing to change in this track
		isHeader, trackChannel := isHeaderTrackAndGetTrackChannel(curTrack)
		if isHeader {
			// we're not doing anything with this track, but we still want it to be included in the list
			tracksWithLoweredVolume = append(tracksWithLoweredVolume, curTrack)
			continue
		}

		// I would have thought that for channel 3, the controlChangeStatus number should be 0xB3, but
		// it seems the MIDIs created for ACC all use 0xB0, 0x80, and 0x90 for all channels
		controlChangeStatus := controlChangeStatusNum // + trackChannel

		// Get all midi events via iterator
		iter := curTrack.GetIterator()
		volumeMIDIEvent := createNewVolumeEvent(curTrack, nonEmphasizedTrackVolume, trackChannel)

		var newTrack *smf.Track

		var eventPos = uint32(0)
		for iter.MoveNext() {
			// grab the track name so we can name our output files correctly
			if iter.GetValue().GetMetaType() == smf.MetaSequenceTrackName {
				trackNameMap[currentTrackNum] = grabTrackName(iter.GetValue())
			}
			// once we've found the MIDI event that's setting the channel volume, replace the old MIDI event with one that has the desired channel volume
			if iter.GetValue().GetStatus() == controlChangeStatus && iter.GetValue().GetData()[0] == volumeControllerNum {
				// fmt.Println(iter.GetValue().String())
				// fmt.Println(volumeMIDIEvent.String())
				newTrack = createNewTrack(curTrack, eventPos, volumeMIDIEvent, currentTrackNum)
				tracksWithLoweredVolume = append(tracksWithLoweredVolume, newTrack)
				break
			}
			eventPos++
		}

	}

	var newMIDIFilesToBeCreated []*smf.MIDIFile
	var emphasizedTrackNum = uint16(0)
	for i := 0; i < len(tracksAtFullVolume); i++ {
		// Create division
		division, err := smf.NewDivision(960, smf.NOSMTPE)
		if err != nil {
			log.Printf("Failed to create new Division object with error: %v", err)
		}

		// Create new midi struct
		newMIDIFile, err := smf.NewSMF(smf.Format1, *division)
		if err != nil {
			log.Printf("Failed to create new MIDI object with error: %v", err)
		}

		// var fullVolTrack *smf.Track
		fullVolTrack := tracksAtFullVolume[emphasizedTrackNum]
		for k := 0; k < len(tracksWithLoweredVolume); k++ {
			if uint16(k) == emphasizedTrackNum {
				newMIDIFile.AddTrack(fullVolTrack)
			} else {
				newMIDIFile.AddTrack(tracksWithLoweredVolume[k])
			}
			// createMIDIFile(newTrack, currentTrackNum, tracksWithLoweredVolume, trackName)
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

func writeNewMIDIFile(wg *sync.WaitGroup, fileNum int, newMidiFile *smf.MIDIFile, trackNameMap map[uint16]string, midiFileName string) {
	defer wg.Done()
	// Save to new midi file
	var newFileName string
	if trackName, ok := trackNameMap[uint16(fileNum)]; ok {
		newFileName = "./output/" + midiFileName + "_" + trackName + ".mid"
	} else {
		return
	}

	fmt.Println("Creating", newFileName)
	outputMidi, err := os.Create(newFileName)
	if err != nil {
		log.Panicf("Failed to create new MIDI file with error: %v", err)
	}
	defer outputMidi.Close()

	writer := bufio.NewWriter(outputMidi)
	smfio.Write(writer, newMidiFile)
	writer.Flush()
}

func grabTrackName(e smf.Event) string {
	var trackName string
	for _, c := range e.GetData() {
		char := fmt.Sprintf("%c", c)
		trackName += char
	}
	return trackName
}

// checks if there are any MIDI events in the track
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

func createNewTrack(track *smf.Track, replacePos uint32, newEvent *smf.MIDIEvent, currentTrackNum uint16) *smf.Track {
	allTrackEvents := track.GetAllEvents()
	// replace volume-setting MIDI event with our lowered-volume event
	allTrackEvents[replacePos] = newEvent

	// create a new track with our updated array of events
	updatedTrack, err := smf.TrackFromArray(allTrackEvents)
	if err != nil {
		log.Panicf("Failed to create new track from event list with error: %v", err)
	}

	return updatedTrack
}

func createNewVolumeEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatusNum, channel, volumeControllerNum, newVolume)
	if err != nil {
		log.Panicf("Failed to create new MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}
