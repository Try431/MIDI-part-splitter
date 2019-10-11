#!/bin/bash

source_dir=$1
dest_dir=$2
for file in "$source_dir"/*.mid; do
	FILENAME=$(echo $file | cut -d . -f 1)
	MIDI_FILE=$FILENAME.mid
	WAV_FILE=$FILENAME.wav
	MP3_FILE=$FILENAME.mp3
	fluidsynth -F $WAV_FILE /usr/share/sounds/sf2/FluidR3_GM.sf2 $MIDI_FILE
	ffmpeg -i $WAV_FILE -vn -ar 44100 -ac 2 -b:a 320k $MP3_FILE
    rm $WAV_FILE
done
dest_dir_char_count=${#dest_dir}
num_to_print=$((37+$dest_dir_char_count))
echo
printf '%.0s%%' $(seq $num_to_print)
echo
echo "%%%%%%% Storing MP3 files in $dest_dir %%%%%%%"
printf '%.0s%%' $(seq $num_to_print)
echo
mkdir -p "$dest_dir" && mv "$source_dir"/*.mp3 "$dest_dir"
echo "All done! 😄"