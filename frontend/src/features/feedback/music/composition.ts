import { random,type MusicProfile } from './profile';
export const BPM=56;
export const BEAT_SECONDS=60/BPM;
export type MusicNote={midi:number;instrument:'pad'|'bell'|'pluck'|'wind'|'crystal'|'bass'|'flute'|'mallet'|'chip'|'gust'|'surf'|'rustle';offset:number;duration:number;gain:number;pan:number};
const chords=[[60,64,67,71],[60,65,69,72],[62,67,71,74],[60,64,69,72],[62,65,69,72],[64,67,71,74]];
// Eight harmonic steps span the entire four-phrase arc, including a return home.
const progressions={spacious:[0,1,3,5,1,4,2,0],organic:[0,3,1,0,4,1,2,0],flowing:[0,5,3,1,4,3,2,0],kinetic:[0,2,3,5,1,4,2,0],reflective:[3,5,1,0,4,1,2,0]};
const rhythms=[[0,3,6,8,11,14],[0,2,5,8,10,13],[0,3,7,8,12,14],[0,2,6,9,11,14]];
const scale=[67,69,71,72,74,76,77,79,81,83,84];
function modal(note:number,profile:MusicProfile) {
 const pitch=note%12;
 return note-(profile.mode!=='major'&&(pitch===4||pitch===11)?1:profile.mode==='minor'&&pitch===9?1:0);
}
function chordAt(beat:number,profile:MusicProfile) {
 const arc=Math.floor(beat/64),step=Math.floor(beat/8)%8;
 const progression=progressions[profile.mood];
 // Related substitutions refresh later arcs without losing the final cadence.
 const substitutions=[5,4,2,3,1,0];
 const vary=random(profile.seed^Math.imul(arc+1,97)^step)();
 const index=arc>0&&step>0&&step<6&&vary<.4?substitutions[progression[step]]:progression[step];
 const variants=[[0,1,2,3,4,5],[3,4,5,0,1,2],[1,3,4,5,2,0]];
 return chords[variants[profile.harmony][index]].map(note=>modal(note,profile));
}
function melody(phrase:number,profile:MusicProfile) {
 const section=phrase%4,arc=Math.floor(phrase/4),rng=random(profile.seed^Math.imul(arc+1,0x85ebca6b));
 const rhythm=rhythms[(profile.rhythm+Math.floor(rng()*rhythms.length))%rhythms.length];
 const count=section===3?2:Math.min(6,3+Math.round(profile.density*4)+(section===2?1:0));
 const positions=section===3?[0,6]:rhythm.slice(0,count);
 let previous=section===1?76:72;
 return positions.map((position,i)=>{
  const beat=phrase*16+position,chord=chordAt(beat,profile);
  const motifIndex=section===1?profile.motif.length-1-i%profile.motif.length:i%profile.motif.length;
  const target=scale[profile.motif[motifIndex]+2]+(section===2?2:section===1?-2:0);
  // Strong beats and phrase endings resolve into harmony; passing notes move by step.
  const resolve=position%4===0||i===positions.length-1;
  const candidates=scale.map(note=>modal(note,profile)).filter(note=>!resolve||chord.some(tone=>tone%12===note%12));
  const midi=candidates.reduce((best,note)=>{
   const score=(n:number)=>Math.abs(n-target)+Math.max(0,Math.abs(n-previous)-4)*2;
   return score(note)<score(best)?note:best;
  },candidates[0]);
  previous=midi;
  return{position,midi,index:i};
 });
}
export function voiceLead(previous:number[],palette:number[]) {
 let best:number[]=[60,64,67],cost=Infinity;
 const candidates=[...new Set(palette.flatMap(note=>[note-12,note,note+12]))].filter(note=>note>=60&&note<=79).sort((a,b)=>a-b);
 for(let a=0;a<candidates.length;a++)for(let b=a+1;b<candidates.length;b++)for(let c=b+1;c<candidates.length;c++){
  const next=[candidates[a],candidates[b],candidates[c]];
  if(next[1]-next[0]<2||next[2]-next[1]<2||next[2]-next[0]>16)continue;
  const score=next.reduce((sum,note,i)=>sum+Math.abs(note-(previous[i]??[60,64,67][i])),0)+Math.max(0,next[0]-65)*.5;
  if(score<cost){cost=score;best=next;}
 }
 return best;
}
export function phraseNotes(beat:number,profile:MusicProfile,previousVoicing:number[]):{notes:MusicNote[];voicing:number[]} {
 const phrase=Math.floor(beat/16),section=phrase%4,rng=random(profile.seed^Math.imul(phrase+1,0x9e3779b1)^Math.imul(beat+1,73));
 const chord=chordAt(beat,profile);
 let voicing=previousVoicing;const notes:MusicNote[]=[];
 // Four phrases form an arc: establish, answer, develop, then leave room to breathe.
 if(beat%8===0 && !(section===3&&beat%16===8)){
  voicing=voiceLead(previousVoicing,chord);
  voicing.forEach((midi,i)=>notes.push({midi,instrument:'pad',offset:i*.14,duration:8.8+rng()*.8,gain:.042+(section===2?.005:0),pan:(i-1)*profile.spread*.55}));
 }
 const entry=melody(phrase,profile).find(note=>note.position===beat%16);
 if(entry){
  const instrument=profile.lead;
  const syncopation=section===1&&entry.index%2===1?.5:0;
  notes.push({midi:entry.midi,instrument,offset:.12+syncopation*BEAT_SECONDS+rng()*.04,duration:instrument==='bell'?3.4:2.4,gain:(section===3?.025:.033)+rng()*.005,pan:(rng()-.5)*profile.spread});
  if(section===2&&entry.index%3===1){
   const upper=chord.map(tone=>tone+12).filter(tone=>tone<=84);
   const midi=upper.reduce((best,tone)=>Math.abs(tone-entry.midi-4)<Math.abs(best-entry.midi-4)?tone:best,upper[0]);
   notes.push({midi,instrument:instrument==='bell'?'mallet':'bell',offset:BEAT_SECONDS*.75,duration:2.4,gain:.020,pan:profile.spread*.3});
  }
 }
 return{notes,voicing};
}
export function textureNotes(beat:number,profile:MusicProfile):MusicNote[] {
 const phrase=Math.floor(beat/16),section=phrase%4,position=beat%16;
 const rng=random(profile.seed^Math.imul(phrase+1,0x27d4eb2d));
 const airy=profile.mood==='spacious'||profile.mood==='flowing';
 const notes:MusicNote[]=[];
 // A single, slowly swelling gust per phrase; the quieter phrase keeps its space.
 if(position===10&&section!==3){
  notes.push({midi:60,instrument:'wind',offset:.3+rng()*.6,duration:6+rng()*1.5,gain:airy?.026:.018,pan:(rng()-.5)*profile.spread});
 }
 // Incidental layers use their own stream so extra ambience cannot alter musical decisions.
 const incident=random(profile.seed^Math.imul(phrase+1,0x165667b1));
 const entry=2+Math.floor(incident()*6),chance=incident();
 if(section!==3&&position===entry&&chance<(airy?.8:.5)){
  const instrument=profile.mood==='flowing'?'surf':profile.mood==='organic'?'rustle':'gust';
  notes.push({midi:60,instrument,offset:.2+incident()*.7,duration:instrument==='gust'?4.2:6.5,
   gain:.018+incident()*.008,pan:(incident()-.5)*profile.spread*1.6});
 }
 // Paired crystalline glints answer the melody, always tuned to the current chord.
 if(position===(section%2===0?12:4)){
  const chord=chordAt(beat,profile),tones=chord.map(note=>note+12).filter(note=>note<=84);
  const index=Math.floor(rng()*tones.length),pan=(rng()-.5)*profile.spread;
  for(let i=0;i<2;i++)notes.push({midi:tones[(index+i)%tones.length],instrument:'crystal',offset:.25+i*BEAT_SECONDS*.5,duration:2.6,gain:(section===3?.012:.019)*(i===0?1:.7),pan:i===0?pan:-pan});
 }
 // Root foundation and a rhythmic inner voice give each palette an audible silhouette.
 const chord=chordAt(beat,profile);
 if(section!==3&&position%8===0){
  notes.push({midi:chord[0]-12,instrument:'bass',offset:.02,duration:3.8,gain:.032,pan:0});
 }
 const stride=profile.motion===2?2:4;
 if(section!==3&&position%stride===stride-1){
  const index=(Math.floor(position/stride)+profile.rhythm+section)%chord.length;
  notes.push({midi:chord[index],instrument:profile.accompaniment,offset:BEAT_SECONDS*(profile.motion===2?.25:.5),duration:1.6,gain:section===2?.023:.017,pan:-profile.spread*.4});
 }
 return notes;
}
export class MusicalTransition {
 private pending?:{profile:MusicProfile;atBeat:number};
 constructor(public current:MusicProfile){}
 request(profile:MusicProfile,seconds:number){
  if(profile.seed===this.current.seed){this.pending=undefined;return;}
  if(this.pending?.profile.seed===profile.seed)return;
  this.pending={profile,atBeat:Math.ceil((seconds/BEAT_SECONDS+.3)/4)*4};
 }
 at(beat:number){if(this.pending&&beat>=this.pending.atBeat){this.current=this.pending.profile;this.pending=undefined;}return this.current;}
}
