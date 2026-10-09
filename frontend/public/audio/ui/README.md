# Gemu spatial UI audio

Original deterministic synthesis created for this project; no samples or third-party music. Source: `scripts/generate-ui-audio.py`. Stereo PCM WAV, 48 kHz / 16 bit, soft envelopes and peaks below −6 dBFS. All effects are project assets for redistribution with Gemu. No external attribution or runtime service required. This notice describes provenance; the repository owner determines the project's distribution license.

The palette uses detuned glass-like partials, a warm low register, subtle filtered air, stereo multi-tap echoes and a diffuse room (damped comb filters and allpass diffusion). Reverb and delay are baked into the WAVs; runtime playback requires no convolution or effects network. Routine effects use a rounded 35 ms attack, subdued upper partials and a fading tail. Navigation uses E4 within the background C-major palette, with lower intensity than confirmations. Panel and navigation effects do not duck the music; explicit confirmations use a gentle 18% reduction with gradual recovery.

| Effect | Duration including tail | Character |
| --- | --- | --- |
| `welcome` | 3.920 s | Layered, low-to-high fifth/octave motif; spacious stereo echoes and long diffuse decay |
| `navigate` | 0.610 s | Soft E4 focus tone with diffuse room and quiet reflections |
| `section` | 0.715 s | Open, warm section shift |
| `open` | 0.940 s | Ascending panel signal |
| `close` | 0.790 s | Descending panel signal |
| `favorite-on` | 0.940 s | Bright two-note confirmation |
| `favorite-off` | 0.680 s | Soft single-note release |
| `confirm` | 0.980 s | Spacious explicit sound test |
| `error` | 0.600 s | Low, non-alarming signal with restrained room |

Nine files total 1,954,004 bytes. Regenerate with `python3 scripts/generate-ui-audio.py` from the repository root. Adjust loudness after listening on TV and headphones. Natural welcome completion lets its tail finish over the library; skipping, muting or losing window focus fades out active voices.

## Procedural background composition

Background music is now synthesized locally by `frontend/src/features/feedback/music/`. The title, subtitle, genre and description determine a local text-derived profile and melodic motif. Shared major harmony, phrase development and breathing space replace the repeating 24-second recording. No external samples or musical API are used.

Game changes preserve transport, active notes and the common stereo reverb/echo. This procedural music uses its own synthesis/effects graph; the nine UI WAV effects above remain baked and independent. The legacy `ambient.wav` is no longer distributed. `scripts/generate-ambient-audio.py` is retained only as a historical experiment.
