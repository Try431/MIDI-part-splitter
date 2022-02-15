import midi2audio
import pydub
from os import path, remove
import boto3
import json
import urllib.parse


S3_CLIENT = boto3.client('s3')
COMPONENT_MIDI_FILES_BUCKET = "component-midi-files"
CREATED_MP3_FILES_BUCKET = "created-mp3-files"

CURRENT_DIR = path.dirname(__file__)
FLUIDSYNTH_FILE_LOC = path.join(CURRENT_DIR, './FluidR3_GM.sf2')
WAV_TEMP_FILE_LOC = "/tmp/temp_wav_file.wav"
# MP3_OUTPUT_FILE_LOC = "/tmp/<name_of_original_mid_file>.mp3"
# MP3_OUTPUT_FILE_LOC = "/tmp/mp3_output_file.mp3"

# Note that /tmp is capped at 512 MB, so will want to clear stuff up as I'm done using it

# the key here will eventually be passed in via the SQS message payload
def grab_midi_from_s3(key=None):
    resp = S3_CLIENT.download_file(Bucket=COMPONENT_MIDI_FILES_BUCKET,
                                   Key=key,
                                   Filename=f"/tmp/{key}")
    
    # print(resp)
    
def convert_midi_to_mp3(key=None):
    fs = midi2audio.FluidSynth(sound_font=FLUIDSYNTH_FILE_LOC)
    fs.midi_to_audio(midi_file=f"/tmp/{key}", audio_file=WAV_TEMP_FILE_LOC)
    
    # clean up MIDI file, as we've already used it to create an audio file
    remove(f"/tmp/{key}")

    wav_audio = pydub.AudioSegment.from_file(WAV_TEMP_FILE_LOC, format="wav")
    mp3_file = key.replace(".mid", "") + ".mp3"
    wav_audio.export(f"/tmp/{mp3_file}", format="mp3")
    
    print(f"Uploading mp3 file {key} to s3")
    try:
        S3_CLIENT.upload_file(f"/tmp/{mp3_file}", CREATED_MP3_FILES_BUCKET, mp3_file)
    except Exception as e:
        print(f"Failed to upload /tmp/{mp3_file} to {CREATED_MP3_FILES_BUCKET} with exception {e}")
        raise e
        
    # clean up mp3 files, as we've uploaded them to s3
    if path.exists(WAV_TEMP_FILE_LOC):
        print("Cleaning up unnecessary intermediate files")
        remove(WAV_TEMP_FILE_LOC)
        
    if path.exists(f"/tmp/{mp3_file}"):
        print("Cleaning up mp3 files")
        remove(f"/tmp/{mp3_file}")


def handler(event=None, context=None):
    print("Received event payload")
    bodies = []
    component_midi_keys = []
    for r in event.get("Records"):
        body = json.loads(r.get("body"))
        bodies.append(body)
        
    for b in bodies:
        for record in b.get("Records"):
            key = record.get("s3").get("object").get("key")
            parsed_key = urllib.parse.unquote(key)
            component_midi_keys.append(parsed_key)
            
    print("component_midi_keys")
    print(component_midi_keys)
    for key in component_midi_keys:
        grab_midi_from_s3(key)
        convert_midi_to_mp3(key)
        

if __name__ == '__main__':
    handler()