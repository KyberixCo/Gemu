import type { Game, GameMetadata } from '../../../catalog';
export type MusicMood = 'spacious'|'organic'|'flowing'|'kinetic'|'reflective';
export type MusicProfile = { seed:number; mood:MusicMood; density:number; brightness:number; spread:number; motif:number[]; harmony:number; rhythm:number; motion:1|2; mode:'major'|'minor'|'dorian'; lead:'bell'|'pluck'|'flute'|'chip'; accompaniment:'mallet'|'chip'|'pluck' };
export const normalizeText=(text:string)=>text.normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLowerCase();
export function textSeed(text:string) {
 let hash=2166136261;for(const char of text){hash^=char.charCodeAt(0);hash=Math.imul(hash,16777619);}return hash>>>0;
}
export function random(seed:number) {
 let value=seed>>>0;return()=>{value+=0x6D2B79F5;let n=value;n=Math.imul(n^(n>>>15),n|1);n^=n+Math.imul(n^(n>>>7),n|61);return((n^(n>>>14))>>>0)/4294967296;};
}
const words:Record<MusicMood,string[]>={
 spacious:['exploracion','astral','universo','planeta','orbita','frontera','espacio','space','exploration'],
 organic:['valle','hogar','bosque','sendero','naturaleza','wild','valley','forest','aventura'],
 flowing:['oceano','marea','tide','isla','costa','mar','water','sea'],
 kinetic:['carreras','circuito','velocidad','curva','drift','neon','arcade','maniobra','racing'],
 reflective:['recuerdo','historia','reino','ceniza','luz','echoes','intima','decision','rpg','story'],
};
const parameters:Record<MusicMood,{density:number;brightness:number;spread:number}>={
 spacious:{density:.24,brightness:1600,spread:.7},organic:{density:.42,brightness:1900,spread:.4},
 flowing:{density:.38,brightness:1800,spread:.6},kinetic:{density:.62,brightness:2300,spread:.45},reflective:{density:.28,brightness:1500,spread:.5},
};
// Local bilingual keyword analysis + text-derived motif. No external AI service or API.
export function profileForGame(game:Pick<Game,'title'|'subtitle'|'genre'|'description'> & GameMetadata):MusicProfile {
 const text=normalizeText(`${game.title} ${game.subtitle} ${game.genre} ${game.description} ${(game.tags??[]).map(normalizeText).sort().join(' ')}`);
 const genre=normalizeText(game.genre);const tokens=text.split(/[^a-z0-9]+/);
 let mood:MusicMood='spacious',best=0;
 for(const [candidate,vocabulary] of Object.entries(words)){
  const score=vocabulary.reduce((sum,word)=>sum+(tokens.some(token=>token.startsWith(word))?1:0)+(genre.includes(word)?2:0),0);
  if(score>best){best=score;mood=candidate as MusicMood;}
 }
 // Separate title and metadata streams: naming changes rhythm/harmony, not just a few pitches.
 const titleSeed=textSeed(normalizeText(game.title));
 const seed=textSeed(`${text}|${normalizeText(game.platform??'unknown')}|${game.releaseYear??'unknown'}|${game.emulated??false}`);
 const rng=random(seed),identity=random(titleSeed);
 const retro=game.emulated===true||(game.releaseYear!==undefined&&game.releaseYear<1996);
 const mode=mood==='reflective'?'minor':mood==='spacious'||mood==='flowing'?'dorian':'major';
 const lead=retro?'chip':mood==='organic'?'flute':mood==='kinetic'?'pluck':'bell';
 return{seed,mood,...parameters[mood],brightness:retro?3200:parameters[mood].brightness,
  harmony:Math.floor(identity()*3),rhythm:Math.floor(identity()*4),motion:mood==='kinetic'||retro?2:1,mode,lead,
  accompaniment:retro?'chip':mood==='organic'||mood==='reflective'?'mallet':'pluck',
  motif:Array.from({length:8},()=>Math.floor(rng()*7))};
}
export const neutralProfile=profileForGame({title:'Gemu',subtitle:'Todos tus mundos',genre:'Exploración',description:'Luz y aire'});
