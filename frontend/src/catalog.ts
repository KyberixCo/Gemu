export type GameArtwork = {id:string;type:'procedural';variant:0|1|2}|{id:string;type:'image';src:string;position?:string};
export type GameMetadata = { releaseYear?:number; platform?:string; emulated?:boolean; tags?:string[];installed?:boolean;cover?:string;logo?:string;icon?:string;player?:import('./library').PlayerState };
export type Game = { id: string; title: string; subtitle: string; genre: string; provider: string; color: string; symbol: string; description: string;artworks?:GameArtwork[] } & GameMetadata;
const defaultArtwork:GameArtwork={id:'primary',type:'procedural',variant:0};
export function artworksForGame(game:Game):GameArtwork[]{return game.artworks?.length?game.artworks:[defaultArtwork];}
// Original procedural artwork: no copyrighted covers or external network dependency.
export const games: Game[] = [
 { id:'astral',releaseYear:2024,platform:'PC',emulated:false,tags:['science fiction','open world'],title:'Astral Frontier',subtitle:'Más allá de lo conocido',genre:'Exploración',provider:'Steam',color:'#627f91',symbol:'◉',description:'Un universo silencioso. Una señal lejana. Traza tu propio camino entre planetas olvidados y descubre qué se esconde al otro lado de la frontera.' },
 { id:'drift',releaseYear:2023,platform:'PC',emulated:false,tags:['synthwave','time trial'],title:'Neon Drift',subtitle:'La noche es tu circuito',genre:'Carreras',provider:'Steam',color:'#b167ae',symbol:'／',description:'Domina las curvas de una ciudad eléctrica. Cada vuelta es una oportunidad para superar tu mejor tiempo.' },
 { id:'wild',releaseYear:2022,platform:'PC',emulated:false,tags:['cozy','crafting'],title:'Wild Valley',subtitle:'Encuentra tu lugar',genre:'Aventura',provider:'Epic Games',color:'#81996d',symbol:'△',description:'Explora un valle vivo, descubre sus senderos y construye un nuevo hogar a tu ritmo.' },
 { id:'echo',releaseYear:2021,platform:'PC',emulated:false,tags:['dark fantasy','soulslike'],title:'Echoes of Ash',subtitle:'Una última luz',genre:'Acción RPG',provider:'Steam',color:'#ae795a',symbol:'✳',description:'Recorre un reino de cenizas y protege la última luz. Una aventura desafiante donde cada decisión deja huella.' },
 { id:'orbit',releaseYear:1992,platform:'Mega Drive',emulated:true,tags:['retro','shoot em up'],title:'Orbit Zero',subtitle:'Sin margen de error',genre:'Arcade',provider:'Manual',color:'#d0c36b',symbol:'◎',description:'Pilota entre órbitas imposibles y encadena maniobras en esta experiencia arcade de precisión.' },
 { id:'tide',releaseYear:2019,platform:'PC',emulated:false,tags:['narrative','ocean'],title:'Low Tide',subtitle:'El océano recuerda',genre:'Indie',provider:'Steam',color:'#5e9c9b',symbol:'≈',description:'Una aventura íntima entre islas y recuerdos. Sigue la marea y escucha las historias que dejó atrás.' },
].map(game=>({...game,artworks:[{id:'primary',type:'procedural',variant:0},{id:'horizon',type:'procedural',variant:1},{id:'dusk',type:'procedural',variant:2}] as GameArtwork[]}));
export function filterGames(catalog: Game[], query: string, source: string, favoritesOnly: boolean, favorites: string[]) {
 const normalize = (s: string) => s.normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLowerCase();
 return catalog.filter(g => normalize(`${g.title} ${g.genre}`).includes(normalize(query)) && (source === 'Todos' || g.provider === source) && (!favoritesOnly || favorites.includes(g.id)));
}
