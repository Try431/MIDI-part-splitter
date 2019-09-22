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

const volumeControllerNum = uint8(0x07)

func main() {

	// Open test midi file
	file, _ := os.Open("./assets/dominefiliunigenite_high_sop.mid")
	defer file.Close()

	// Read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))
	if err != nil {
		log.Panicf("Failed to read MIDI file %v with error: %v", file, err)
	}

	var currentTrack uint16
	for currentTrack = 0; currentTrack < midi.GetTracksNum(); currentTrack++ {
		track := midi.GetTrack(currentTrack)
		// if there is no MIDI_EVENT in the track, there's nothing to change in this track
		if isHeaderTrack(track) {
			continue
		}

		// TESTING
		if currentTrack == 2 {
			break
		}

		trackChannel := track.GetAllEvents()[0].GetChannel()
		controlChangeStatus := 0xB0 + trackChannel

		// Get all midi events via iterator
		iter := track.GetIterator()
		volumeMIDIEvent := createNewVolumeEvent(track, 75, trackChannel)

		var eventPos = uint32(0)
		var newTrack *smf.Track
		var trackName string
		for iter.MoveNext() {
			if iter.GetValue().GetMetaType() == smf.MetaSequenceTrackName {
				trackName = grabTrackName(iter.GetValue())
			}
			fmt.Println(iter.GetValue())
			// TESTING
			if iter.GetValue().GetStatus() == controlChangeStatus && iter.GetValue().GetData()[0] == volumeControllerNum {
				newTrack = replaceEvent(track, eventPos, volumeMIDIEvent)
				break
			}
			eventPos++
		}

		fmt.Println(newTrack)
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
		// fmt.Println(e.GetData())
		// fmt.Println(e.String())
		if strings.HasPrefix(e.String(), "MIDI") {
			headerEvent = false
			break
		}
	}
	return headerEvent
}

func replaceEvent(track *smf.Track, replacePos uint32, newEvent *smf.MIDIEvent) *smf.Track {
	allTrackEvents := track.GetAllEvents()
	allTrackEvents[replacePos] = newEvent

	updatedTrack, err := smf.TrackFromArray(allTrackEvents)
	if err != nil {
		log.Printf("Failed to create new track from event list with error: %v", err)
	}

	return updatedTrack
}

func createNewVolumeEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	var controlChangeStatus uint8
	controlChangeStatus = 0xB0 + channel
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatus, channel, volumeControllerNum, newVolume)
	if err != nil {
		log.Printf("Failed to create new MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}
