package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
)

const volumeControllerNum = uint8(0x7)

func main() {

	// Open test midi file
	file, _ := os.Open("./assets/dominefiliunigenite_high_sop.mid")
	defer file.Close()

	// Read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))

	if err != nil {
		fmt.Println(err)
	}

	// Print number of midi tracks
	// fmt.Println(midi.GetTracksNum())

	// Get zero track
	track := midi.GetTrack(0)

	// Get all midi events via iterator
	iter := track.GetIterator()

	for iter.MoveNext() {
		// fmt.Println(iter.GetValue())
	}

	track1 := midi.GetTrack(1)
	iter1 := track1.GetIterator()

	volumeMIDIEvent := createNewVolumeEvent(track1, 75)
	fmt.Println(volumeMIDIEvent)
	var i uint32
	i = 0
	for iter1.MoveNext() {
		fmt.Println(iter1.GetValue())
		if iter1.GetValue().GetStatus() == 0xB0 && iter1.GetValue().GetData()[0] == 0x07 {
			fmt.Println(iter1.GetValue().GetChannel())
			// fmt.Println("Data:", iter1.GetValue().GetData())
			// newTrack := replaceEvent(track1, i, volumeMIDIEvent)
			// fmt.Println("Event after at cur pos:", newTrack.GetAllEvents()[i])
			break
		}
		i++
	}

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

func createNewVolumeEvent(t *smf.Track, newVolume uint8) *smf.MIDIEvent {
	trackChannel := t.GetAllEvents()[0].GetChannel()
	var controlChangeStatus uint8
	controlChangeStatus = 0xB0 + trackChannel
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatus, trackChannel, volumeControllerNum, newVolume)
	if err != nil {
		log.Printf("Failed to create new MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}
