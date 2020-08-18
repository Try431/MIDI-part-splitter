#!/bin/bash

source_file=$1
dest_dir=$2
silent=$3
FILENAME=.$(echo $source_file | cut -d . -f 2)
MIDI_FILE=$FILENAME.mid
WAV_FILE=$FILENAME.wav
MP3_FILE=$FILENAME.mp3
fluidsynth -F $WAV_FILE convert/FluidR3_GM.sf2 $MIDI_FILE >/dev/null
if [[ $silent = "false" ]]; then
    echo "Converting '$source_file' to an MP3 file in '$dest_dir'" 
fi
ffmpeg -y -i $WAV_FILE -vn -ar 44100 -ac 2 -b:a 320k $MP3_FILE
rm $WAV_FILE

mkdir -p "$dest_dir" && mv "$MP3_FILE" "$dest_dir"