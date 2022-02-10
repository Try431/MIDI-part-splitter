# MIDI-part-splitter

## Dependencies
Debian/Ubuntu
```
sudo apt install -y fluidsynth ffmpeg
```


## Installation
Clone repository and build

````
$ git clone git@github.com:Try431/MIDI-part-splitter.git
$ cd ./MIDI-part-splitter
$ go build
````

## Goals

- [x] Create emphasized MIDI files
- [x] Set emphasized track to a different MIDI instrument
- [x] Create .mp3 files from MIDI files
- [ ] Build a GUI for using the tool


## How to use

### View flag options

````
$ ./MIDI-part-splitter -h
Usage of ./MIDI-part-splitter:
  -d string
    	Directory containing .mid files you wish to parse - will recursively search subdirectories
    	(e.g., './MIDI-part-splitter -d ./dir/to/search/')
  -f string
    	Name of .mid file you wish to parse
    	(e.g., './MIDI-part-splitter -f midi_file.mid')
  -inst int
    	Instrument number for emphasized track - see README for instrument list
    	(e.g., './MIDI-part-splitter -f midi_file.mid -inst 22)  (default 65)
  -o string
    	Directory where mp3 files will be stored
    	(e.g., './MIDI-part-splitter -f midi_file.mid -o ./dir/to/store/mp3s) (default "./output/mp3s")
  -vol int
    	Volume of de-emphasized voice tracks - must be between 0 and 100
    	(e.g., './MIDI-part-splitter -f midi_file.mid -vol 30) (default 40)
````

### Examples 
Given an `assets` directory containing the following MIDI files:


````
$ ls assets/
dominefiliunigenite.mid cumsanctospiritu.mid
````

#### Example 1 - Parsing a single MIDI file

````
$ ./MIDI-part-splitter -f ./assets/dominefiliunigenite.mid
Creating ./output/dominefiliunigenite_Soprano.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Alto.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Tenor.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Bass.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Piano.mid with all other tracks set to volume 40
````

#### Example 2 - Parsing a single MIDI file, and setting the volume of the non-emphasized tracks

````
$ ./MIDI-part-splitter -f ./assets/cumsanctospiritu.mid -vol 20
Creating ./output/dominefiliunigenite_Soprano.mid with all other tracks set to volume 20
Creating ./output/dominefiliunigenite_Alto.mid with all other tracks set to volume 20
Creating ./output/dominefiliunigenite_Tenor.mid with all other tracks set to volume 20
Creating ./output/dominefiliunigenite_Bass.mid with all other tracks set to volume 20
Creating ./output/dominefiliunigenite_Piano.mid with all other tracks set to volume 20
````

#### Example 3 - Parsing all MIDI files in a directory

````
$ ./MIDI-part-splitter -d ./assets/
Creating ./output/dominefiliunigenite_Soprano.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Alto.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Tenor.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Bass.mid with all other tracks set to volume 40
Creating ./output/dominefiliunigenite_Piano.mid with all other tracks set to volume 40
Creating ./output/cumsanctospiritu_Soprano.mid with all other tracks set to volume 40
Creating ./output/cumsanctospiritu_Alto.mid with all other tracks set to volume 40
Creating ./output/cumsanctospiritu_Tenor.mid with all other tracks set to volume 40
Creating ./output/cumsanctospiritu_Bass.mid with all other tracks set to volume 40
Creating ./output/cumsanctospiritu_Piano.mid with all other tracks set to volume 40
````

## Instrument List

| Code Number  | Instrument Name  |
|---|---|
| 0 | Acoustic Grand Piano |
| 1 | Bright Acoustic Piano |
| 2 | Electric Grand Piano |
| 3 | Honky-tonk Piano |
| 4 | Electric Piano 1 |
| 5 | Electric Piano 2 |
| 6 | Harpsichord |
| 7 | Clavi |
| 8 | Celesta |
| 9 | Glockenspiel |
| 10 | Music Box |
| 11 | Vibraphone |
| 12 | Marimba |
| 13 | Xylophone |
| 14 | Tubular Bells |
| 15 | Dulcimer |
| 16 | Drawbar Organ |
| 17 | Percussive Organ |
| 18 | Rock Organ |
| 19 | Church Organ |
| 20 | Reed Organ |
| 21 | Accordion |
| 22 | Harmonica |
| 23 | Tango Accordion |
| 24 | Acoustic Guitar (nylon) |
| 25 | Acoustic Guitar (steel) |
| 26 | Electric Guitar (jazz) |
| 27 | Electric Guitar (clean) |
| 28 | Electric Guitar (muted) |
| 29 | Overdriven Guitar |
| 30 | Distortion Guitar |
| 31 | Guitar harmonics |
| 32 | Acoustic Bass |
| 33 | Electric Bass (finger) |
| 34 | Electric Bass (pick) |
| 35 | Fretless Bass |
| 36 | Slap Bass 1 |
| 37 | Slap Bass 2 |
| 38 | Synth Bass 1 |
| 39 | Synth Bass 2 |
| 40 | Violin |
| 41 | Viola |
| 42 | Cello |
| 43 | Contrabass |
| 44 | Tremolo Strings |
| 45 | Pizzicato Strings |
| 46 | Orchestral Harp |
| 47 | Timpani |
| 48 | String Ensemble 1 |
| 49 | String Ensemble 2 |
| 50 | Synth Strings 1 |
| 51 | Synth Strings 2 |
| 52 | Choir Aahs |
| 53 | Voice Oohs |
| 54 | Synth Voice |
| 55 | Orchestra Hit |
| 56 | Trumpet |
| 57 | Trombone |
| 58 | Tuba |
| 59 | Muted Trumpet |
| 60 | French Horn |
| 61 | Brass Section |
| 62 | Synth Brass 1 |
| 63 | Synth Brass 2 |
| 64 | Soprano Sax |
| 65 | Alto Sax |
| 66 | Tenor Sax |
| 67 | Baritone Sax |
| 68 | Oboe |
| 69 | English Horn |
| 70 | Bassoon |
| 71 | Clarinet |
| 72 | Piccolo |
| 73 | Flute |
| 74 | Recorder |
| 75 | Pan Flute |
| 76 | Blown Bottle |
| 77 | Shakuhachi |
| 78 | Whistle |
| 79 | Ocarina |
| 80 | Lead 1 (square) |
| 81 | Lead 2 (sawtooth) |
| 82 | Lead 3 (calliope) |
| 83 | Lead 4 (chiff) |
| 84 | Lead 5 (charang) |
| 85 | Lead 6 (voice) |
| 86 | Lead 7 (fifths) |
| 87 | Lead 8 (bass + lead) |
| 88 | Pad 1 (new age) |
| 89 | Pad 2 (warm) |
| 90 | Pad 3 (polysynth) |
| 91 | Pad 4 (choir) |
| 92 | Pad 5 (bowed)|
| 93 | Pad 6 (metallic) |
| 94 | Pad 7 (halo) |
| 95 | Pad 8 (sweep) |
| 96 | FX 1 (rain) |
| 97 | FX 2 (soundtrack) |
| 98 | FX 3 (crystal) |
| 99 | FX 4 (atmosphere) |
| 100 | FX 5 (brightness) |
| 101 | FX 6 (goblins) |
| 102 | FX 7 (echoes) |
| 103 | FX 8 (sci-fi) |
| 104 | Sitar |
| 105 | Banjo |
| 106 | Shamisen |
| 107 | Koto |
| 108 | Kalimba |
| 109 | Bag Pipe |
| 110 | Fiddle |
| 111 | Shanai |
| 112 | Tinkle Bell |
| 113 | Agogo |
| 114 | Steel Drums |
| 115 | Woodblock |
| 116 | Taiko Drum |
| 117 | Melodic Tom |
| 118 | Synth Drum |
| 119 | Reverse Cymbal |
| 120 | Guitar Fret Noise |
| 121 | Breath Noise |
| 122 | Seashore |
| 123 | Bird Tweet |
| 124 | Telephone Ring |
| 125 | Helicopter |
| 126 | Applause |
| 127 | Gunshot |


## Built With

* [Go](https://golang.org/)
* [Bash](https://www.gnu.org/software/bash/)
* [WebAssembly](https://webassembly.org)

## Author

* **Try431** - *main developer* - [Try431](https://github.com/Try431)

## Helpful References for MIDI Technical Specs
- [https://www.csie.ntu.edu.tw/~r92092/ref/midi/](https://www.csie.ntu.edu.tw/~r92092/ref/midi/)
- [https://www.nyu.edu/classes/bello/FMT_files/9_MIDI_code.pdf](https://www.nyu.edu/classes/bello/FMT_files/9_MIDI_code.pdf)
- [https://en.wikipedia.org/wiki/MIDI_timecode](https://en.wikipedia.org/wiki/MIDI_timecode)
- [https://sites.uci.edu/camp2014/2014/05/19/timing-in-midi-files/](https://sites.uci.edu/camp2014/2014/05/19/timing-in-midi-files/)
- [http://www.music-software-development.com/midi-tutorial.html](http://www.music-software-development.com/midi-tutorial.html)
- [https://www.midi.org/specifications/item/gm-level-1-sound-set](https://www.midi.org/specifications/item/gm-level-1-sound-set)