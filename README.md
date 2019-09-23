# MIDI-part-splitter

## Installation
Clone repository and build

````
$ git clone git@github.com:Try431/MIDI-part-splitter.git
$ cd ./MIDI-part-splitter
$ go build
````


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
  -vol int
    	[Optional] Volume of de-emphasized voice tracks - must be between 0 and 100 (default 40)

````
Given an `assets` directory containing the following MIDI files:

### Examples 

````
$ ls assets/
dominefiliunigenite.mid cumsanctospiritu.mid
````

#### Example 1 - Parsing a single MIDI file with SATB split

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


## Built With

* [Go](https://golang.org/) - The Go Programming Language

## Author

* **Try431** - *main developer* - [Try431](https://github.com/Try431)

## Helpful References for MIDI Technical Specs
- https://www.csie.ntu.edu.tw/~r92092/ref/midi/
- https://www.nyu.edu/classes/bello/FMT_files/9_MIDI_code.pdf
- https://en.wikipedia.org/wiki/MIDI_timecode
- https://sites.uci.edu/camp2014/2014/05/19/timing-in-midi-files/
