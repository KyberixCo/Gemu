import { priority, soundNames, SoundPolicy, type SoundName } from './policy';
import type { Game } from '../../catalog';
import { ProceduralMusic } from './music/engine';
import { neutralProfile,profileForGame } from './music/profile';
import { defaultPreferences, type FeedbackPreferences } from './preferences';

type Voice = { source: AudioBufferSourceNode; gain: GainNode; priority: number };
export class InterfaceAudio {
 private context?: AudioContext;
 private unlocked=false;
 private musicUnavailable=false;
 private master?: GainNode;
 private buffers = new Map<SoundName, AudioBuffer>();
 private files = new Map<SoundName, ArrayBuffer>();
 private ambientSession = false;
 private music?: ProceduralMusic;
 private musicProfile=neutralProfile;
 private analyser?: AnalyserNode;
 private analysisData = new Float32Array(512);

 private loading?: Promise<void>;
 private decoding?: Promise<void>;
 private voices: Voice[] = [];
 private policy = new SoundPolicy();
 private preferences: FeedbackPreferences = defaultPreferences;
 private active = true;
 private generation = 0;
 preload() {
  return this.loading ??= Promise.all(soundNames.map(async name => {
   try { const response = await fetch(`/audio/ui/${name}.wav`); if (response.ok) this.files.set(name, await response.arrayBuffer()); } catch { /* Missing audio is a silent degradation. */ }
  })).then(() => undefined);
 }
 configure(preferences: FeedbackPreferences) {
  this.preferences = preferences;
  if (!preferences.sound || preferences.volume === 0) this.stopEffects();
  if (this.context && this.master) {
   const now = this.context.currentTime;
   this.master.gain.cancelScheduledValues(now);
   this.master.gain.setTargetAtTime(preferences.sound ? preferences.volume / 100 : 0, now, .008);
  }
  this.syncAmbient();
 }
 async unlock(): Promise<boolean> {
  try {
   if (!this.context) {
    this.context = new AudioContext();
    this.master = this.context.createGain();
    this.master.gain.value = this.preferences.sound ? this.preferences.volume / 100 : 0;
    this.master.connect(this.context.destination);
   }
   // Call resume within the originating gesture, before awaiting file loading.
   const resumed = this.context.resume();
   this.decoding ??= this.preload().then(async () => {
    await Promise.all([...this.files].map(async ([name, file]) => {
     try { this.buffers.set(name, await this.context!.decodeAudioData(file.slice(0))); } catch { /* Unsupported file stays silent. */ }
    }));
    this.files.clear();
   });
   await Promise.all([resumed, this.decoding]);
   this.unlocked=this.context.state==='running';
   this.syncAmbient();
   return this.context.state === 'running';
  } catch { return false; }
 }
 play(name: SoundName, changed = true) {
  const context = this.context, buffer = this.buffers.get(name);
  const enabled = this.active && this.preferences.sound && this.preferences.volume > 0 && (name !== 'welcome' || this.preferences.welcomeSound);
  if (!context || context.state !== 'running' || !buffer || !this.master || !this.policy.accept(name, performance.now(), enabled, changed)) return false;
  const level = priority(name);
  if (level > 0) this.voices.filter(v => v.priority === 0).forEach(v => this.remove(v));
  if (this.voices.length >= 2) {
   const weakest = [...this.voices].sort((a,b) => a.priority - b.priority)[0];
   if (weakest.priority > level) return false;
   this.remove(weakest);
  }
  const source = context.createBufferSource(), gain = context.createGain();
  source.buffer = buffer; source.connect(gain); gain.connect(this.master);
  const voice = { source, gain, priority: level };
  this.voices.push(voice);
  source.onended = () => { this.voices = this.voices.filter(v => v !== voice); source.disconnect(); gain.disconnect(); };
  source.start();
  if (name === 'welcome' || name === 'confirm' || name === 'error') this.duckAmbient(Math.min(buffer.duration,.8));
  return true;
 }
 private remove(voice: Voice) {
  if (!this.context) return;
  const now = this.context.currentTime;
  voice.gain.gain.cancelScheduledValues(now);
  voice.gain.gain.setTargetAtTime(0, now, .005);
  try { voice.source.stop(now + .025); } catch { /* Already ended. */ }
  this.voices = this.voices.filter(v => v !== voice);
 }
 cancelPending() { this.generation++; }
 private stopEffects() { this.cancelPending(); [...this.voices].forEach(v => this.remove(v)); this.policy.reset(); }
 stop() { this.stopEffects(); this.stopAmbient(); }
 setActive(active: boolean) { this.active = active; if (!active) this.stop(); else this.syncAmbient(); }
 setAmbientSession(enabled: boolean) { this.ambientSession = enabled; this.syncAmbient(); }
 selectGame(game:Pick<Game,'title'|'subtitle'|'genre'|'description'> & Pick<Game,'releaseYear'|'platform'|'emulated'|'tags'>) {
  const profile=profileForGame(game);
  if(profile.seed===this.musicProfile.seed)return;
  this.musicProfile=profile;this.music?.select(profile);
 }
 private syncAmbient() {
  const context=this.context;
  const enabled=this.active&&this.ambientSession&&this.preferences.ambient&&this.preferences.ambientVolume>0;
  if(!enabled){this.stopAmbient();return;}
  if(!context||!this.unlocked||context.state!=='running'||this.musicUnavailable)return;
  try{
   if(!this.music){
    this.analyser=context.createAnalyser();this.analyser.fftSize=512;this.analyser.connect(context.destination);
    this.music=new ProceduralMusic(context,this.analyser,this.musicProfile);
   }
   this.music.setVolume(this.preferences.ambientVolume/100);this.music.start();
  }catch{this.musicUnavailable=true;this.analyser?.disconnect();this.analyser=undefined;/* UI effects remain operational. */}
 }
 private stopAmbient(){this.music?.stop();}
 private duckAmbient(duration:number){this.music?.duck(duration);}
 ambientFrame() {
  const music=this.music?.frame()??{playing:false,seconds:0,pulse:0,crystal:0,wind:0,warmth:0,pan:0};let level=0;
  if(music.playing&&this.analyser){
   this.analyser.getFloatTimeDomainData(this.analysisData);
   let energy=0;for(const sample of this.analysisData)energy+=sample*sample;
   level=Math.min(1,Math.sqrt(energy/this.analysisData.length)*100);
  }
  return{...music,level};
 }
 // Used only for explicit welcome/test actions. Never replay old navigation after unlock.
 async preview(name: 'welcome' | 'confirm'): Promise<boolean> {
  const generation = this.generation;
  const unlocked = await this.unlock();
  return unlocked && generation === this.generation ? this.play(name) : false;
 }
}
export const interfaceAudio = new InterfaceAudio();
