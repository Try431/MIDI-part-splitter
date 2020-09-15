from midi2audio import FluidSynth
from pydub import AudioSegment

fs = FluidSynth(sound_font="convert/FluidR3_GM.sf2")
output_wav = "convert/test_output.wav"
fs.midi_to_audio(midi_file="convert/Hine_Ma_Tov_Bass.mid", audio_file=output_wav)

wav_audio = AudioSegment.from_file(output_wav, format="wav")
wav_audio.export("convert/test_output.mp3", format="mp3")