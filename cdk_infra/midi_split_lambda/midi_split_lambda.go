package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Try431/EasyMIDI/smf"
	"github.com/Try431/EasyMIDI/smfio"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	MIDI_FILE_DROPOFF_BUCKET    = "midi-file-dropoff"
	COMPONENT_MIDI_FILES_BUCKET = "component-midi-files"
)

// 0xCn is the code for setting a program change command for channel n
// a program change is used solely to change between different instruments/presets/patches, depending on the device
const programChangeStatusNum = uint8(0xC0)

// 0xBn is the code for setting a control change command for channel n
// a control change can be used to set a variety of settings and functions; in this code, I use it solely for setting main channel volume
const controlChangeStatusNum = uint8(0xB0)

// 0x07 is the code for a control change number for a channel's main volume
const volumeControllerNum = uint8(0x07)

// MIDIOutputDirectory the directory where the converted MIDI files will be stored
var MIDIOutputDirectory = "/tmp"

// NonEmphasizedTrackVolume the volume at which to set the non-emphasized tracks (default 40)
var NonEmphasizedTrackVolume = uint8(40)

// EmphasizedTrackVolume we must set the emphasized track volume to 100 because some MIDI tracks have non-100 default volumes
const EmphasizedTrackVolume = uint8(100)

// EmphasizedInstrumentNum the number corresponding to the instrument played by the emphasized track (default 65: alto sax)
var EmphasizedInstrumentNum = uint8(65)

// MP3OutputDirectory the directory where the mp3 files will be stored (default output/mp3s)
var MP3OutputDirectory = "output/mp3s"

// SilenceOutput when true, effectively stops all output to stdout
var SilenceOutput = false

// outputMIDIFilePaths is a slice of the full filepaths of the MIDI files created by writeNewMIDIFile() -- this slice will be accessed by the conversion bash script
var (
	outputMIDIFilePaths []string
	filepathLock        sync.RWMutex
)

func printWrapper(toPrint string) {
	if !SilenceOutput {
		fmt.Println(toPrint)
	}
}

type MIDILambdaPayload struct {
	MIDIFilenames           []string `json:"midi_filenames"`
	NonEmphasizedVolume     *int     `json:"non_emphasized_volume,omitempty"`
	EmphasizedInstrumentNum *int     `json:"emphasized_instrument_num,omitempty"`
}

type SQSEventPayload struct {
	Records []struct {
		Body string `json:"body"`
	} `json:"Records"`
}

type S3EventPayload struct {
	Records []struct {
		EventVersion string    `json:"eventVersion"`
		EventSource  string    `json:"eventSource"`
		AwsRegion    string    `json:"awsRegion"`
		EventTime    time.Time `json:"eventTime"`
		EventName    string    `json:"eventName"`
		UserIdentity struct {
			PrincipalID string `json:"principalId"`
		} `json:"userIdentity"`
		RequestParameters struct {
			SourceIPAddress string `json:"sourceIPAddress"`
		} `json:"requestParameters"`
		ResponseElements struct {
			XAmzRequestID string `json:"x-amz-request-id"`
			XAmzID2       string `json:"x-amz-id-2"`
		} `json:"responseElements"`
		S3 struct {
			S3SchemaVersion string `json:"s3SchemaVersion"`
			ConfigurationID string `json:"configurationId"`
			Bucket          struct {
				Name          string `json:"name"`
				OwnerIdentity struct {
					PrincipalID string `json:"principalId"`
				} `json:"ownerIdentity"`
				Arn string `json:"arn"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Size      int    `json:"size"`
				ETag      string `json:"eTag"`
				Sequencer string `json:"sequencer"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func HandleRequest(ctx context.Context, payload SQSEventPayload) error {
	fmt.Printf("Received payload: %v\n", payload)
	var s3Payload S3EventPayload

	err := json.Unmarshal([]byte(payload.Records[0].Body), &s3Payload)
	if err != nil {
		fmt.Println("Failed to unmarshal S3 payload into json struct with error: ", err)
		return err
	}

	midiFilename := s3Payload.Records[0].S3.Object.Key
	downloadFromS3Bucket(midiFilename)
	fmt.Println("Finished downloading all files in payload")

	fmt.Println("Starting MIDI split process...")

	var wg sync.WaitGroup // this isn't really useful right now, but adding so I don't have to modify SplitParts func
	wg.Add(1)
	fPath := "/tmp/" + midiFilename
	SplitParts(&wg, fPath)
	wg.Wait()

	// if payload.NonEmphasizedVolume != nil {
	// 	NonEmphasizedTrackVolume = uint8(*payload.NonEmphasizedVolume)
	// }

	// if payload.EmphasizedInstrumentNum != nil {
	// 	EmphasizedInstrumentNum = uint8(*payload.EmphasizedInstrumentNum)
	// }

	// if len(payload.MIDIFilenames) != 0 {
	// 	// TODO - look into downloading in parallel
	// 	for _, midiFilename := range payload.MIDIFilenames {
	// 		downloadFromS3Bucket(midiFilename)
	// 	}
	// 	fmt.Println("Finished downloading all files in payload")
	// } else {
	// 	fmt.Println("No filenames provided in payload - exiting.")
	// }

	// fmt.Println("Starting MIDI split process...")

	// var wg sync.WaitGroup
	// wg.Add(len(payload.MIDIFilenames))

	// for i := 0; i < len(payload.MIDIFilenames); i++ {
	// 	fPath := "/tmp/" + payload.MIDIFilenames[i]
	// 	SplitParts(&wg, fPath)
	// }
	// wg.Wait()
	fmt.Println("All done!")

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}

func downloadFromS3Bucket(midiFilename string) {
	localFileLoc := "/tmp/" + midiFilename

	file, err := os.Create(localFileLoc)
	if err != nil {
		fmt.Println(err)
	}

	defer file.Close()

	sess, _ := session.NewSession(&aws.Config{
		// TODO - potentially make this into an env variable?
		Region: aws.String("us-east-2"),
	},
	)

	downloader := s3manager.NewDownloader(sess)

	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(MIDI_FILE_DROPOFF_BUCKET),
			Key:    aws.String(midiFilename),
		})
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")
}

// SplitParts splits the MIDI file into different voice parts and creates new MIDI files
// with those voice parts emphasized
func SplitParts(mainWg *sync.WaitGroup, midiFilePath string) {
	outputMIDIFilePaths = []string{}
	defer mainWg.Done()
	file, err := os.Open(midiFilePath)
	midiFileName := strings.TrimSuffix(filepath.Base(midiFilePath), filepath.Ext(midiFilePath))
	if err != nil {
		log.Fatalf("Failed to open %v with error: %v", midiFilePath, err)
	}
	defer file.Close()

	// read and save midi to smf.MIDIFile struct
	midi, err := smfio.Read(bufio.NewReader(file))
	if err != nil {
		log.Fatalf("Failed to read MIDI file %v with error: %v", file, err)
	}

	// collecting record of all tracks in the MIDI file so we can construct our new MIDI files in the same track order
	var tracksWithLoweredVolume []*smf.Track
	var tracksAtFullVolume []*smf.Track
	trackNameMap := make(map[uint16]string)

	// iterating through all tracks in MIDI file
	for currentTrackNum := uint16(0); currentTrackNum < midi.GetTracksNum(); currentTrackNum++ {
		curTrack := midi.GetTrack(currentTrackNum)
		isHeader, trackChannel := isHeaderTrackAndGetTrackChannel(curTrack)

		// if there is no MIDI_EVENT in the track (i.e., is a header track which consists solely of META_EVENTs), there's nothing to change in this track
		if isHeader {
			// we're not doing anything with this track, but we still want it to be included in the list because of indexing purposes when creating our output midi files
			tracksAtFullVolume = append(tracksAtFullVolume, curTrack)
			tracksWithLoweredVolume = append(tracksWithLoweredVolume, curTrack)
			continue
		}
		eventPos := uint32(0)
		newInsIter := curTrack.GetIterator()
		newInstrumentEvent := createNewInstrumentEvent(curTrack, EmphasizedInstrumentNum, trackChannel)
		var newInstrumentTrack *smf.Track
		// create a new track with the emphasized instrument
		for newInsIter.MoveNext() {
			if newInsIter.GetValue().GetStatus() == programChangeStatusNum {
				newInstrumentTrack = createNewTrack(curTrack, eventPos, newInstrumentEvent)
				break
			}
			eventPos++
		}
		// create an emphasized track with the new instrument and at full volume
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

		// if a track didn't end up having a name, we're going to give it a generic name
		if _, ok := trackNameMap[currentTrackNum]; !ok {
			genericTrackName := "autogenerated_name_track_" + strconv.Itoa(int(currentTrackNum))
			trackNameMap[currentTrackNum] = genericTrackName
		}
	}

	var newMIDIFilesToBeCreated []*smf.MIDIFile
	emphasizedTrackNum := uint16(0)
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

	fmt.Println("Finished creating new midi files")

	err = uploadMIDIFilesToS3(outputMIDIFilePaths)
	if err != nil {
		fmt.Println(err)
	}
}

func uploadMIDIFilesToS3(filenames []string) error {
	// The session the S3 Uploader will use
	sess := session.Must(session.NewSession())

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)
	for _, filename := range filenames {

		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("failed to open file %q, %v", filename, err)
		}

		trimmedFilename := strings.ReplaceAll(filename, "/tmp/", "")
		// Upload the file to S3.
		result, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(COMPONENT_MIDI_FILES_BUCKET),
			Key:    aws.String(trimmedFilename),
			Body:   f,
		})
		if err != nil {
			return fmt.Errorf("failed to upload file, %v", err)
		}
		fmt.Printf("file uploaded to %s\n", aws.StringValue(&result.Location))
	}

	return nil
}

func runConversionScript(wg *sync.WaitGroup, filepath string) {
	defer wg.Done()
	cmd := exec.Command("/bin/bash", "convert/convert_async.sh", filepath, MP3OutputDirectory, strconv.FormatBool(SilenceOutput))
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
		newFileName = MIDIOutputDirectory + "/" + midiFileName + "_" + trackName + ".mid"
	} else {
		return
	}

	fmt.Println("Creating ", newFileName, " with all other tracks set to volume ", NonEmphasizedTrackVolume)

	// newpath := filepath.Join(".", MIDIOutputDirectory)
	// err := os.MkdirAll(newpath, os.ModePerm)
	// if err != nil {
	// 	log.Fatalf("Failed to create directory %v with error: %v", newFileName, err)
	// }
	outputMidi, err := os.Create(newFileName)
	if err != nil {
		log.Fatalf("Failed to create new MIDI file with error: %v", err)
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
	// replace all slashes with underscores
	trackName = strings.ReplaceAll(trackName, "/", "_")
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
	pos := uint32(0)
	volCount := 0
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
		log.Fatalf("Failed to create new track from event list with error: %v", err)
	}

	return updatedTrack
}

// Creates a new MIDI_EVENT to set the volume of the track we want to de-emphasize
func createNewVolumeEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	newVolumeMIDIEvent, err := smf.NewMIDIEvent(0, controlChangeStatusNum, channel, volumeControllerNum, newVolume)
	if err != nil {
		log.Fatalf("Failed to create new volume MIDI event with error: %v", err)
	}
	return newVolumeMIDIEvent
}

func createNewInstrumentEvent(t *smf.Track, newVolume uint8, channel uint8) *smf.MIDIEvent {
	newInstrumentEvent, err := smf.NewMIDIEvent(0, programChangeStatusNum, channel, EmphasizedInstrumentNum, 0)
	if err != nil {
		log.Fatalf("Failed to create new instrument MIDI event with error: %v", err)
	}
	return newInstrumentEvent
}
