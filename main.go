package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
)

const controlChangeStatusNum = uint8(0xB0)
const volumeControllerNum = uint8(0x07)

var nonEmphasizedTrackVolume = uint8(40)

func main() {

	// Open test midi file
	file, _ := os.Open("./assets/dominefiliunigenite.mid")
	defer file.Close()

	// Read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))
	if err != nil {
		log.Panicf("Failed to read MIDI file %v with error: %v", file, err)
	}

	// Collecting record of all tracks in the MIDI file so we can construct our new MIDI files in the same track order
	var tracksWithLoweredVolume []*smf.Track
	for k := uint16(0); k < midi.GetTracksNum(); k++ {
		tracksWithLoweredVolume = append(tracksWithLoweredVolume, midi.GetTrack(k))
	}

	// iterating through all tracks in MIDI file
	for currentTrackNum := uint16(0); currentTrackNum < midi.GetTracksNum(); currentTrackNum++ {
		curTrack := midi.GetTrack(currentTrackNum)
		// if there is no MIDI_EVENT in the track, there's nothing to change in this track
		if isHeaderTrack(curTrack) {
			continue
		}

		trackChannel := curTrack.GetAllEvents()[0].GetChannel()
		controlChangeStatus := controlChangeStatusNum + trackChannel

		// Get all midi events via iterator
		iter := curTrack.GetIterator()
		volumeMIDIEvent := createNewVolumeEvent(curTrack, nonEmphasizedTrackVolume, trackChannel)

		var newTrack *smf.Track

		var eventPos = uint32(0)
		var trackName string
		for iter.MoveNext() {
			// grab the track name so we can name our output files correctly
			if iter.GetValue().GetMetaType() == smf.MetaSequenceTrackName {
				trackName = grabTrackName(iter.GetValue())
			}
			// fmt.Println(iter.GetValue())
			// once we've found the MIDI event that's setting the channel volume, replace the old MIDI event with one that has the desired channel volume
			if iter.GetValue().GetStatus() == controlChangeStatus && iter.GetValue().GetData()[0] == volumeControllerNum {
				fmt.Println(iter.GetValue().String())
				fmt.Println(volumeMIDIEvent.String())
				newTrack = createNewTrack(curTrack, eventPos, volumeMIDIEvent, currentTrackNum, tracksWithLoweredVolume)
				createMIDIFile(newTrack, currentTrackNum, tracksWithLoweredVolume, trackName)
				break
			}
			eventPos++
		}

		if currentTrackNum == 1 {
			break
		}
	}

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
func isHeaderTrack(track *smf.Track) bool {
	allEvents := track.GetAllEvents()
	headerEvent := true
	for _, e := range allEvents {
		// fmt.Println(e.String())
		if strings.HasPrefix(e.String(), "MIDI") {
			headerEvent = false
			break
		}
	}
	return headerEvent
}

func createNewTrack(track *smf.Track, replacePos uint32, newEvent *smf.MIDIEvent, currentTrackNum uint16, allTracks []*smf.Track) *smf.Track {
	allTrackEvents := track.GetAllEvents()
	allTrackEvents[replacePos] = newEvent

	updatedTrack, err := smf.TrackFromArray(allTrackEvents)
	if err != nil {
		log.Printf("Failed to create new track from event list with error: %v", err)
	}

	return updatedTrack
}

func createMIDIFile(newTrack *smf.Track, trackToUpdate uint16, allTracks []*smf.Track, filename string) {
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

	for i := uint16(0); i < uint16(len(allTracks)); i++ {
		if i == trackToUpdate {
			newMIDIFile.AddTrack(newTrack)
		} else {
			newMIDIFile.AddTrack(allTracks[i])
		}
	}

	// Save to new midi source file
	outputMidi, err := os.Create(filename + ".mid")
	if err != nil {
		log.Panicf("Failed to create new MIDI file with error: %v", err)
	}
	defer outputMidi.Close()

	// Create buffering stream
	writer := bufio.NewWriter(outputMidi)
	smfio.Write(writer, newMIDIFile)
	writer.Flush()
}

func createNewVolumeEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	controlChangeStatus := controlChangeStatusNum + channel
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatus, channel, volumeControllerNum, newVolume)
	if err != nil {
		log.Printf("Failed to create new MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}
