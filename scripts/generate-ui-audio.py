"""Original Gemu spatial UI synthesis. Stdlib only; no external samples.

Detuned glass partials + airy filtered noise + stereo multi-tap echo + a
Schroeder-style diffuse room. All processing is baked into the local WAVs.
"""
from pathlib import Path
import math
import random
import struct
import wave

ROOT = Path(__file__).resolve().parents[1] / 'frontend/public/audio/ui'
RATE = 48000
# name: (notes [start, duration, frequency, amplitude, stereo pan], tail, wet, echo)
SOUNDS = {
    'welcome': ([(0, 1.15, 130.81, .13, -.2), (.12, 1.1, 196, .09, .25),
                 (.32, 1.25, 261.63, .10, -.35), (.58, 1.15, 392, .075, .3),
                 (.82, 1.1, 523.25, .055, -.15)], 2.0, .58, .32),
    'navigate': ([(0, .19, 329.63, .045, 0)], .42, .48, .08),
    'section': ([(0, .16, 261.63, .07, -.2), (.045, .17, 392, .045, .2)], .5, .40, .19),
    'open': ([(0, .18, 329.63, .07, -.25), (.07, .22, 493.88, .045, .25)], .65, .46, .23),
    'close': ([(0, .16, 392, .055, .2), (.06, .18, 261.63, .04, -.2)], .55, .42, .2),
    'favorite-on': ([(0, .19, 523.25, .055, -.2), (.08, .21, 659.25, .04, .2)], .65, .46, .22),
    'favorite-off': ([(0, .18, 392, .055, 0)], .5, .38, .18),
    'confirm': ([(0, .20, 329.63, .07, -.2), (.04, .24, 493.88, .045, .2)], .7, .48, .24),
    'error': ([(0, .20, 174.61, .075, 0)], .4, .28, .1),
}


def synthesize(notes, tail, welcome, seed):
    size = math.ceil((max(start + length for start, length, *_ in notes) + tail) * RATE)
    left, right = [0.0] * size, [0.0] * size
    rng = random.Random(seed)
    for start, length, hz, amplitude, pan in notes:
        air = 0.0
        for i in range(round(length * RATE)):
            t = i / RATE
            attack = min(1, t / (.055 if welcome else .035))
            release = min(1, max(0, (length - t) / (.3 if welcome else .10)))
            envelope = attack * release * math.exp(-(2.2 if welcome else 3.0) * t / length)
            air += .09 * (rng.uniform(-1, 1) - air)
            for channel, detune, balance in ((left, -.003, math.sqrt((1-pan)/2)),
                                             (right, .003, math.sqrt((1+pan)/2))):
                base = hz * (1 + detune)
                # Slightly inharmonic upper partials provide a glass-like, floating tone.
                body = (math.sin(2*math.pi*base*t) + (.19 if welcome else .07)*math.sin(2*math.pi*base*(2.01 if welcome else 2)*t)*math.exp(-5*t/length)
                        + (.075 if welcome else .012)*math.sin(2*math.pi*base*(3.97 if welcome else 4)*t)*math.exp(-8*t/length))
                value = amplitude * balance * envelope * (body + .055*air)
                channel[round(start*RATE)+i] += value
    return left, right


def room(signal, tail, channel):
    # Parallel damped combs, then serial allpasses to diffuse distinct reflections.
    wet = [0.0] * len(signal)
    delays = (.0297, .0371, .0411, .0437) if channel == 0 else (.0313, .0353, .0397, .0451)
    for seconds in delays:
        delay = round(seconds * RATE)
        ring, cursor, damped = [0.0] * delay, 0, 0.0
        feedback = 10 ** (-3 * seconds / max(.18, tail*.72))
        for i, sample in enumerate(signal):
            reflected = ring[cursor]
            damped += .32 * (reflected-damped)
            ring[cursor] = sample + damped*feedback
            wet[i] += reflected*.25
            cursor = (cursor+1) % delay
    for seconds in (.0047+channel*.0006, .0017+channel*.0002):
        delay = round(seconds*RATE)
        ring, cursor = [0.0]*delay, 0
        for i, sample in enumerate(wet):
            delayed = ring[cursor]
            output = delayed-.55*sample
            ring[cursor] = sample+.55*output
            wet[i] = output
            cursor = (cursor+1) % delay
    return wet


def render(name, notes, tail, wet_mix, echo_mix):
    left, right = synthesize(notes, tail, name == 'welcome', sum(map(ord, name)))
    reflections = [room(left, tail, 0), room(right, tail, 1)]
    output = [[0.0]*len(left), [0.0]*len(left)]
    for c, dry in enumerate((left, right)):
        other = right if c == 0 else left
        echo_delay = .19 if name == 'welcome' else .075
        for i, sample in enumerate(dry):
            value = sample + reflections[c][i]*wet_mix
            for tap in range(1, 5 if name == 'welcome' else 3):
                offset = round((echo_delay*tap + c*.012)*RATE)
                if i >= offset:
                    value += (other if tap % 2 else dry)[i-offset]*echo_mix*(.5**(tap-1))
            # Fade the very end of the baked room; no truncation clicks.
            fade = min(1, (len(dry)-1-i)/(.09*RATE))
            output[c][i] = value*max(0, fade)
    peak = max(abs(v) for channel in output for v in channel)
    # Retain natural relative loudness; attenuate only if the peak would exceed −6 dBFS.
    gain = min(1, .48/peak) if peak else 1
    with wave.open(str(ROOT / f'{name}.wav'), 'wb') as out:
        out.setparams((2, 2, RATE, len(left), 'NONE', 'not compressed'))
        out.writeframes(b''.join(struct.pack('<hh', round(l*gain*32767), round(r*gain*32767))
                                 for l, r in zip(*output)))
    print(f'{name}: {len(left)/RATE:.3f}s, peak {20*math.log10(peak*gain):.1f} dBFS')


if __name__ == '__main__':
    ROOT.mkdir(parents=True, exist_ok=True)
    for name, (notes, tail, wet, echo) in SOUNDS.items():
        render(name, notes, tail, wet, echo)
    print(f'{len(SOUNDS)} original stereo WAVs; {sum((ROOT / f"{name}.wav").stat().st_size for name in SOUNDS):,} bytes.')
