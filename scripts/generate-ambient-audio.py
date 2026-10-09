"""Legacy 24-second experiment; not used by the procedural runtime.

Gemu original seamless ambient field: luminous major harmony, airy choir-like tones and stereo echoes."""
from pathlib import Path
import importlib.util
import math
import struct
import wave

ROOT = Path(__file__).resolve().parents[1]
PROFILE = {'duration':24,'breathPeriod':6,'chimes':[3,9,15,21],'echoDelay':.37,'fadeIn':1.8}
spec = importlib.util.spec_from_file_location('ui_synthesis', ROOT/'scripts/generate-ui-audio.py')
ui = importlib.util.module_from_spec(spec)
spec.loader.exec_module(ui)
RATE = ui.RATE
DURATION = PROFILE['duration']
SIZE = RATE*DURATION
TAU = 2*math.pi


def build():
    channels = [[0.0]*SIZE for _ in range(2)]
    # G3–C4–E4–G4–C5: open C-major harmony, no sub-bass.
    # Integral cycles across the loop ensure a continuous sustained bed.
    tones = [(196, .010), (261.63, .017), (329.63, .012), (392, .009), (523.25, .005)]
    for c, values in enumerate(channels):
        for i in range(SIZE):
            t = i/RATE
            breath = .55+.45*math.cos(TAU*(t-3)/PROFILE['breathPeriod'])
            value = 0.0
            for j, (pitch, level) in enumerate(tones):
                cycles = round(pitch*DURATION)+(c if j > 0 else 0)
                phase = c*.4+j*.65
                choir = math.sin(TAU*cycles*t/DURATION+phase)
                choir += .07*math.sin(TAU*cycles*2*t/DURATION+phase)
                value += choir*level*(.82+.18*breath)
            # Air is a low-level periodic cluster, also seamless.
            value += .00035*breath*(math.sin(TAU*30011*t/DURATION)+math.sin(TAU*41873*t/DURATION))
            values[i] = value
    for cue, pitch in zip(PROFILE['chimes'], (523.25, 659.25, 783.99, 1046.5)):
        notes = [(0, 1.4, pitch, .035, -.25), (.24, 1.3, pitch*1.5, .014, .25)]
        dry = ui.synthesize(notes, 2.2, True, int(cue*71))
        for c in range(2):
            reflected = ui.room(dry[c], 2.2, c)
            for i in range(len(dry[c])):
                value = dry[c][i]*.48+reflected[i]*.8
                for tap in range(1,4):
                    delay = round((PROFILE['echoDelay']*tap+c*.016)*RATE)
                    if i >= delay:
                        value += dry[1-c][i-delay]*.32*(.5**(tap-1))
                # Circular accumulation carries the last reverberation through the seam.
                channels[c][(round(cue*RATE)+i)%SIZE] += value
    peak = max(abs(v) for channel in channels for v in channel)
    assert peak < .5
    output = ROOT/'artifacts/audio/ambient-legacy.wav'
    output.parent.mkdir(parents=True, exist_ok=True)
    with wave.open(str(output),'wb') as out:
        out.setparams((2,2,RATE,SIZE,'NONE','not compressed'))
        out.writeframes(b''.join(struct.pack('<hh',round(l*32767),round(r*32767)) for l,r in zip(*channels)))
    print(f'Original seamless stereo ambience: {DURATION}s, {output.stat().st_size:,} bytes, {20*math.log10(peak):.1f} dBFS peak.')


if __name__ == '__main__':
    build()
