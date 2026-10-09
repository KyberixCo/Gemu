import {BEAT_SECONDS,MusicalTransition,phraseNotes,textureNotes,type MusicNote} from './composition';
import {instrumentBuffers,roomImpulse,isAirTexture} from './synthesis';
import {type MusicProfile} from './profile';

type MusicVoice={source:AudioBufferSourceNode;gain:GainNode;pan:StereoPannerNode;end:number};
export class ProceduralMusic {
 private readonly input:GainNode;
 private readonly filter:BiquadFilterNode;
 private readonly output:GainNode;
 private readonly samples:ReturnType<typeof instrumentBuffers>;
 private readonly transition:MusicalTransition;
 private voices:MusicVoice[]=[];
 private timer?:ReturnType<typeof setInterval>;
 private origin=0;
 private pausedSeconds=0;
 private nextBeat=0;
 private voicing=[60,64,67];
 private volume=.25;
 private duckUntil=0;
 private cues:{seconds:number;strength:number;instrument:MusicNote['instrument'];midi:number;pan:number;duration:number}[]=[];
 private running=false;
 constructor(private context:AudioContext,destination:AudioNode,profile:MusicProfile){
  this.transition=new MusicalTransition(profile);this.samples=instrumentBuffers(context);
  this.input=context.createGain();this.input.gain.value=1.25;this.filter=context.createBiquadFilter();this.filter.type='lowpass';this.filter.frequency.value=profile.brightness;
  this.output=context.createGain();this.output.gain.value=0;this.output.connect(destination);
  this.input.connect(this.filter);
  const dry=context.createGain();dry.gain.value=.52;this.filter.connect(dry);dry.connect(this.output);
  const reverb=context.createConvolver(),wet=context.createGain();reverb.buffer=roomImpulse(context);wet.gain.value=.34;
  this.filter.connect(reverb);reverb.connect(wet);wet.connect(this.output);
  const echo=context.createDelay(1),feedback=context.createGain(),echoWet=context.createGain();
  echo.delayTime.value=.37;feedback.gain.value=.28;echoWet.gain.value=.20;
  this.filter.connect(echo);echo.connect(feedback);feedback.connect(echo);echo.connect(echoWet);echoWet.connect(this.output);
 }
 select(profile:MusicProfile){
  if(this.running)this.transition.request(profile,this.seconds());
  else{this.transition.request(profile,this.pausedSeconds);this.transition.at(Infinity);}
 }
 setVolume(volume:number){
  if(this.volume===volume)return;this.volume=volume;
  if(!this.running)return;
  const now=this.context.currentTime;
  this.output.gain.cancelScheduledValues(now);
  this.output.gain.setTargetAtTime(volume*(now < this.duckUntil ? .82 : 1),now,.1);
  if(now<this.duckUntil)this.output.gain.setTargetAtTime(volume,this.duckUntil,.5);
 }
 start(){
  if(this.running||this.context.state!=='running')return;
  this.running=true;this.origin=this.context.currentTime-this.pausedSeconds;
  this.nextBeat=Math.ceil(this.pausedSeconds/BEAT_SECONDS);
  this.output.gain.cancelScheduledValues(this.context.currentTime);this.output.gain.setValueAtTime(0,this.context.currentTime);
  this.output.gain.setTargetAtTime(this.volume,this.context.currentTime,.6);this.duckUntil=0;
  if(this.pausedSeconds>0){
   const chordBeat=Math.floor(this.pausedSeconds/BEAT_SECONDS/8)*8;
   const phrase=phraseNotes(chordBeat,this.transition.current,this.voicing);
   phrase.notes.filter(note=>note.instrument==='pad').forEach(note=>this.schedule(note,this.context.currentTime+.04));
   this.voicing=phrase.voicing;
  }
  this.tick();this.timer=setInterval(()=>this.tick(),50);
 }
 // Short lookahead: late UI frames skip missed beats rather than emitting a burst.
 tick(){
  if(!this.running)return;
  const now=this.context.currentTime,seconds=this.seconds();
  this.nextBeat=Math.max(this.nextBeat,Math.ceil((seconds-.06)/BEAT_SECONDS));
  while(this.origin+this.nextBeat*BEAT_SECONDS<now+.18){
   const previous=this.transition.current.seed,profile=this.transition.at(this.nextBeat);
   const at=Math.max(now+.025,this.origin+this.nextBeat*BEAT_SECONDS);
   if(previous!==profile.seed)this.filter.frequency.setTargetAtTime(profile.brightness,at,2.4);
   const phrase=phraseNotes(this.nextBeat,profile,this.voicing);this.voicing=phrase.voicing;
   phrase.notes.forEach(note=>this.schedule(note,at));
   textureNotes(this.nextBeat,profile).forEach(note=>this.schedule(note,at));this.nextBeat++;
  }
  this.cues=this.cues.filter(cue=>seconds-cue.seconds<Math.max(4,cue.duration));
 }
 private schedule(note:MusicNote,at:number){
  // A bounded ensemble; spare voices are dropped, existing musical tails are never stolen.
  if(this.voices.length>=24)return;
  const start=at+note.offset,source=this.context.createBufferSource(),gain=this.context.createGain(),pan=this.context.createStereoPanner();
  source.buffer=this.samples[note.instrument];source.loop=note.instrument==='pad';
  source.playbackRate.setValueAtTime(isAirTexture(note.instrument)?1:2**((note.midi-60)/12),start);pan.pan.value=note.pan;
  source.connect(gain);gain.connect(pan);pan.connect(this.input);
  const attack=isAirTexture(note.instrument)?1.2:note.instrument==='pad'?1.5:.06,release=isAirTexture(note.instrument)?1.4:note.instrument==='pad'?2.8:.6,end=start+note.duration;
  gain.gain.setValueAtTime(0,start);gain.gain.linearRampToValueAtTime(note.gain,start+attack);
  gain.gain.setValueAtTime(note.gain,Math.max(start+attack,end-release));gain.gain.linearRampToValueAtTime(0,end);
  const voice={source,gain,pan,end};this.voices.push(voice);
  source.onended=()=>{this.voices=this.voices.filter(item=>item!==voice);source.disconnect();gain.disconnect();pan.disconnect();};
  source.start(start);source.stop(end+.01);
  this.cues.push({seconds:start-this.origin,strength:note.instrument==='crystal'?.4:note.instrument==='bell'?.7:.45,instrument:note.instrument,midi:note.midi,pan:note.pan,duration:note.duration});
 }
 duck(duration:number){
  if(!this.running)return;const now=this.context.currentTime;
  this.duckUntil=Math.max(this.duckUntil,now+duration+.1);
  this.output.gain.cancelScheduledValues(now);this.output.gain.setTargetAtTime(this.volume*.82,now,.18);
  this.output.gain.setTargetAtTime(this.volume,this.duckUntil,.5);
 }
 stop(){
  if(!this.running)return;
  const now=this.context.currentTime;this.pausedSeconds=this.seconds();this.running=false;
  clearInterval(this.timer);this.timer=undefined;
  this.output.gain.cancelScheduledValues(now);this.output.gain.setTargetAtTime(0,now,.005);
  for(const voice of this.voices){try{voice.source.stop(now+.025);}catch{/* Already ended. */}}
  this.voices=[];this.cues=[];
 }
 seconds(){return this.running?Math.max(0,this.context.currentTime-this.origin):this.pausedSeconds;}
 frame(){
  const seconds=this.seconds();let pulse=0,crystal=0,wind=0,warmth=0,weight=0,position=0;
  if(this.running)for(const cue of this.cues){
   const age=seconds-cue.seconds;
   if(age<0)continue;
   if(cue.instrument==='pad'||isAirTexture(cue.instrument)){
    if(age>=cue.duration)continue;
    const envelope=Math.min(1,age/(isAirTexture(cue.instrument)?1.2:1.5),(cue.duration-age)/(isAirTexture(cue.instrument)?1.4:2.8));
    if(isAirTexture(cue.instrument))wind+=envelope;
    else{warmth+=Math.sin((cue.midi%12)*Math.PI/6)*envelope;weight+=envelope;}
    continue;
   }
   for(let echo=0;echo<3;echo++){
    const elapsed=age-echo*.37;
    if(elapsed<0)continue;
    const energy=(1-Math.exp(-elapsed/.045))*Math.exp(-elapsed/.65)*cue.strength*(.5**echo);
    pulse+=energy;
    if(cue.instrument==='crystal'){crystal+=energy;position+=cue.pan*energy;}
   }
  }
  return{playing:this.running,seconds,pulse:Math.min(1,pulse),crystal:Math.min(1,crystal),wind:Math.min(1,wind),warmth:weight?warmth/weight:0,pan:crystal?position/crystal:0};
 }
}
