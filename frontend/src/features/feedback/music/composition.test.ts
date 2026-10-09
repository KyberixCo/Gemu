import {describe,it,expect} from 'vitest';
import {games} from '../../../catalog';
import {profileForGame} from './profile';
import {BEAT_SECONDS,MusicalTransition,phraseNotes,textureNotes,voiceLead} from './composition';
describe('text-derived composition',()=>{
 it('spaces ambient textures and leaves the resting phrase free of wind',()=>{
  for(const game of games){
   const profile=profileForGame(game);
   const collect=(phrase:number)=>Array.from({length:16},(_,beat)=>textureNotes(phrase*16+beat,profile)).flat();
   expect(collect(0).filter(note=>note.instrument==='wind')).toHaveLength(1);
   expect(collect(0).filter(note=>note.instrument==='crystal')).toHaveLength(2);
   expect(collect(3).filter(note=>note.instrument==='wind')).toHaveLength(0);
   expect(collect(0)).toEqual(collect(0));expect(collect(0)).not.toEqual(collect(4));
   for(const note of collect(0)){expect(note.gain).toBeLessThan(.04);expect(Math.abs(note.pan)).toBeLessThanOrEqual(1);}
  }
 });
 it('places sparse incidental ambience with deterministic variation and resting space',()=>{
  const instruments=new Set(['gust','surf','rustle']);
  const heard=new Set<string>();
  for(const game of games){
   const profile=profileForGame(game);
   const collect=()=>Array.from({length:512},(_,beat)=>textureNotes(beat,profile).filter(n=>instruments.has(n.instrument)).map(n=>({...n,beat}))).flat();
   const notes=collect();expect(notes).toEqual(collect());expect(notes.length).toBeGreaterThan(0);
   for(const note of notes){heard.add(note.instrument);expect(Math.floor(note.beat/16)%4).not.toBe(3);expect(note.gain).toBeLessThan(.03);expect(Math.abs(note.pan)).toBeLessThanOrEqual(1);}
   for(let phrase=0;phrase<32;phrase++)expect(notes.filter(n=>Math.floor(n.beat/16)===phrase).length).toBeLessThanOrEqual(1);
  }
  expect([...heard].sort()).toEqual(['gust','rustle','surf']);
 });
 it('extracts distinct motifs and semantic moods from game text',()=>{
  const profiles=games.map(profileForGame);expect(new Set(profiles.map(p=>p.seed)).size).toBe(6);
  expect(profiles.map(p=>p.mood)).toEqual(['spacious','kinetic','organic','reflective','kinetic','flowing']);
  expect(profileForGame(games[0])).toEqual(profileForGame({...games[0],description:games[0].description.toUpperCase()}));
 });
 it('changes audible structure for titles and metadata and accepts missing fields',()=>{
  const game=games[4],original=profileForGame(game);
  const modern=profileForGame({...game,emulated:false,releaseYear:2024,platform:'PC'});
  expect(original.lead).toBe('chip');expect(modern.lead).toBe('pluck');
  const collect=(p:ReturnType<typeof profileForGame>)=>Array.from({length:64},(_,beat)=>[...phraseNotes(beat,p,[60,64,67]).notes,...textureNotes(beat,p)]).flat();
  expect(collect(original)).not.toEqual(collect(modern));
  const names=['Aster','Lumen','Cobalt','Nimbus','Solstice','Ember'].map(title=>profileForGame({...game,title}));
  expect(new Set(names.map(p=>`${p.harmony}/${p.rhythm}`)).size).toBeGreaterThan(3);
  expect(profileForGame({...game,tags:[...(game.tags??[])].reverse()})).toEqual(original);
  expect(profileForGame({title:'Test',subtitle:'',genre:'',description:''}).motif).toHaveLength(8);
  for(const note of collect(original)){expect(note.midi).toBeGreaterThanOrEqual(48);expect(note.midi).toBeLessThanOrEqual(84);}
  expect(new Set(collect(original).map(n=>n.instrument))).toContain('bass');
 });
 it('uses descriptions and genres, not only game IDs',()=>{
  const original=profileForGame(games[0]);const modified=profileForGame({...games[0],genre:'Carreras',description:'Circuito, velocidad, curvas y maniobras.'});
  expect(modified.seed).not.toBe(original.seed);expect(modified.mood).toBe('kinetic');expect(modified.density).toBeGreaterThan(original.density);
 });
 it('evolves phrases and deliberately leaves breathing room',()=>{
  const p=profileForGame(games[1]);const collect=(from:number,to:number)=>Array.from({length:to-from},(_,i)=>phraseNotes(i+from,p,[60,64,67]).notes).flat();
  const first=collect(0,16),development=collect(32,48),rest=collect(48,64);
  expect(first).not.toEqual(development);expect(rest.filter(n=>n.instrument==='pad')).toHaveLength(3);
  expect(development.filter(n=>n.instrument==='pad')).toHaveLength(6);
 });
 it('keeps notes in a luminous register and bounds ensemble size',()=>{
  for(const game of games){let voicing=[60,64,67];const profile=profileForGame(game);
   for(let beat=0;beat<512;beat++){const phrase=phraseNotes(beat,profile,voicing);voicing=phrase.voicing;
    expect(phrase.notes.length).toBeLessThanOrEqual(5);
    for(const note of phrase.notes){expect(note.midi).toBeGreaterThanOrEqual(60);expect(note.midi).toBeLessThanOrEqual(84);expect(note.gain).toBeLessThan(.06);}
   }
  }
 });
 it('gives every phrase a melody, with syncopated answers and quieter endings',()=>{
  for(const game of games){
   const profile=profileForGame(game);
   const collect=(phrase:number)=>Array.from({length:16},(_,i)=>({beat:i,notes:phraseNotes(phrase*16+i,profile,[60,64,67]).notes}))
    .flatMap(({beat,notes})=>notes.filter(note=>note.instrument!=='pad').map(note=>({...note,beat})));
   const opening=collect(0),answer=collect(1),development=collect(2),rest=collect(3);
   expect(opening.length).toBeGreaterThanOrEqual(3);
   expect(opening.some(note=>note.beat%2===1)).toBe(true);
   expect(answer.some(note=>note.offset>BEAT_SECONDS*.5)).toBe(true);
   expect(development.some(note=>note.instrument!==opening[0].instrument)).toBe(true);
   expect(rest).toHaveLength(2);
   expect(Math.max(...rest.map(note=>note.gain))).toBeLessThan(Math.min(...opening.map(note=>note.gain)));
   expect(collect(0)).toEqual(opening);
  }
 });
 it('develops harmony across an entire arc and refreshes subsequent arcs',()=>{
  const profile=profileForGame(games[0]);
  const harmony=(from:number)=>Array.from({length:8},(_,i)=>phraseNotes(from+i*8,profile,[60,64,67]).voicing);
  const first=harmony(0);
  expect(first.slice(0,4)).not.toEqual(first.slice(4));
  expect(Array.from({length:4},(_,i)=>harmony((i+1)*64)).some(arc=>JSON.stringify(arc)!==JSON.stringify(first))).toBe(true);
  // The resting phrase omits its second chord; the next arc restores home.
  expect(phraseNotes(64,profile,[60,64,67]).voicing).toEqual(first[0]);
 });
 it('voice-leads compatible chords rather than jumping octaves',()=>{
  const first=[60,64,67],second=voiceLead(first,[60,65,69,72]);expect(second).toEqual([60,65,69]);
  const third=voiceLead(second,[62,67,71,74]);expect(third.every((note,i)=>Math.abs(note-second[i])<=4)).toBe(true);
 });
});
describe('coherent selection transitions',()=>{
 it('keeps the old profile until a common musical boundary',()=>{
  const old=profileForGame(games[0]),next=profileForGame(games[1]),transition=new MusicalTransition(old);
  transition.request(next,BEAT_SECONDS*1.2);expect(transition.at(3)).toEqual(old);expect(transition.at(4)).toEqual(next);
 });
 it('replaces pending choices rather than queuing every highlighted game',()=>{
  const old=profileForGame(games[0]),one=profileForGame(games[1]),two=profileForGame(games[2]),transition=new MusicalTransition(old);
  transition.request(one,BEAT_SECONDS);transition.request(two,BEAT_SECONDS*2);expect(transition.at(4)).toEqual(two);
 });
 it('cancels a pending change when returning to the current game',()=>{
  const old=profileForGame(games[0]),transition=new MusicalTransition(old);transition.request(profileForGame(games[1]),1);transition.request(old,2);
  expect(transition.at(100)).toEqual(old);
 });
});
