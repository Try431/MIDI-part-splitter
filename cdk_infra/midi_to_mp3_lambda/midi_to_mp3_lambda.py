import midi2audio
import pydub
from os import path, remove
import boto3

S3_CLIENT = boto3.client('s3')
S3_BUCKET = "midi-file-dropoff"

CURRENT_DIR = path.dirname(__file__)
FLUIDSYNTH_FILE_LOC = path.join(CURRENT_DIR, './FluidR3_GM.sf2')
WAV_TEMP_FILE_LOC = "/tmp/temp_wav_file.wav"
# MP3_OUTPUT_FILE_LOC = "/tmp/<name_of_original_mid_file>.mp3"
MP3_OUTPUT_FILE_LOC = "/tmp/mp3_output_file.mp3"

# Note that /tmp is capped at 512 MB, so will want to clear stuff up as I'm done using it

# the key here will eventually be passed in via the SQS message payload
def grab_midi_from_s3(key=None):
    resp = S3_CLIENT.download_file(Bucket=S3_BUCKET,
                                   Key="Los_Peces_Bass.mid",
                                   Filename="/tmp/Los_Peces_Bass_file.mid")
    
    print(resp)
    
def convert_midi_to_mp3():
    fs = midi2audio.FluidSynth(sound_font=FLUIDSYNTH_FILE_LOC)
    fs.midi_to_audio(midi_file="/tmp/Los_Peces_Bass_file.mid", audio_file=WAV_TEMP_FILE_LOC)

    wav_audio = pydub.AudioSegment.from_file(WAV_TEMP_FILE_LOC, format="wav")
    wav_audio.export(MP3_OUTPUT_FILE_LOC, format="mp3")

    if path.exists(WAV_TEMP_FILE_LOC):
        # print("Uploading mp3 file to s3")
        # S3_CLIENT.upload_file(WAV_TEMP_FILE_LOC, S3_BUCKET, "Los_Peces_Bass_music.mp3")
        print("Cleaning up unnecessary intermediate files")
        remove(WAV_TEMP_FILE_LOC)
    else:
        print("Can not delete the file as it doesn't exist")


def handler(event=None, context=None):
    grab_midi_from_s3()
    convert_midi_to_mp3()

if __name__ == '__main__':
    handler()