import { describe, it, expect } from 'vitest';
import { games, filterGames,artworksForGame } from './catalog';
describe('library filtering',()=>{
 it('combines source, query and favorites',()=>expect(filterGames(games,'astral','Steam',true,['astral']).map(g=>g.id)).toEqual(['astral']));
 it('ignores accents and case',()=>expect(filterGames(games,'EXPLORACION','Todos',false,[])).toHaveLength(1));
 it('does not show nonfavorites',()=>expect(filterGames(games,'','Todos',true,[])).toEqual([]));
});
describe('title artwork galleries',()=>{
 it('provides distinct procedural variants for every demo title',()=>{
  for(const game of games){const arts=artworksForGame(game);
   expect(arts).toHaveLength(3);expect(new Set(arts.map(art=>art.id)).size).toBe(3);
   expect(arts.map(art=>art.type==='procedural'?art.variant:null)).toEqual([0,1,2]);
  }
 });
 it('accepts multiple image assets and preserves their framing',()=>{
  const artworks=[{id:'landscape',type:'image' as const,src:'/art/landscape.webp',position:'70% center'},{id:'city',type:'image' as const,src:'/art/city.webp'}];
  expect(artworksForGame({...games[0],artworks})).toEqual(artworks);
 });
 it('keeps legacy titles and empty galleries renderable',()=>{
  for(const artworks of [undefined,[]])expect(artworksForGame({...games[0],artworks})).toEqual([{id:'primary',type:'procedural',variant:0}]);
 });
});
