import { useEffect, useRef, useState } from 'react';
import {artworksForGame,type Game,type GameArtwork} from '../../catalog';
import { Scene } from '../../Scene';
import { interfaceAudio } from './audio';
import { sceneLighting } from './scene-lighting';
import {prepareBackgroundArtwork} from '../../background-artwork';
type ArtworkLayer={game:Game;artwork:GameArtwork};
const layerKey=(layer:ArtworkLayer)=>`${layer.game.id}-${layer.artwork.id}`;
export function WorldBackdrop({ game, reduced, enabled, active }: { game: Game; reduced: boolean;enabled:boolean;active:boolean }) {
 const root=useRef<HTMLDivElement>(null);
 useEffect(()=>{
  const element=root.current;if(!element||!enabled||!active||reduced)return;
  let frame=0,last=0,light=sceneLighting({playing:false,level:0,pulse:0,crystal:0,wind:0,warmth:0,pan:0});
  const draw=(now:number)=>{
   if(now-last>=50){
    const delta=last?Math.min((now-last)/1000,.2):.05;last=now;
    const target=sceneLighting(interfaceAudio.ambientFrame()),blend=1-Math.exp(-delta/.22);
    for(const key of Object.keys(light) as (keyof typeof light)[]){
     light[key]+=(target[key]-light[key])*blend;
     element.style.setProperty(`--music-${key}`,String(light[key])+(key==='hue'?'deg':key==='position'?'%':''));
    }
   }
   frame=requestAnimationFrame(draw);
  };
  frame=requestAnimationFrame(draw);
  return()=>{cancelAnimationFrame(frame);element.removeAttribute('style');};
 },[enabled,active,reduced]);
 const [layers,setLayers] = useState<{current:ArtworkLayer;previous:ArtworkLayer|null}>({current:{game,artwork:artworksForGame(game)[0]},previous:null});
 useEffect(() => {
  if (game.id === layers.current.game.id && JSON.stringify(artworksForGame(game)) === JSON.stringify(artworksForGame(layers.current.game))) return;
  let cancelled=false;
  const timer = setTimeout(async()=>{
   let artwork:GameArtwork={id:'fallback',type:'procedural',variant:0};
   for(const candidate of artworksForGame(game)){
    if(await prepareBackgroundArtwork(candidate)){artwork=candidate;break;}
    if(cancelled)return;
   }
   if(!cancelled)setLayers(previous=>({current:{game,artwork},previous:previous.current}));
  },reduced?0:120);
  return()=>{cancelled=true;clearTimeout(timer);};
 },[game,layers.current.game.id,reduced]);
 useEffect(()=>{
  if(!enabled||!active||reduced||layers.previous||game.id!==layers.current.game.id)return;
  const artworks=artworksForGame(game);if(artworks.length<2)return;
  let cancelled=false;
  const timer=setTimeout(async()=>{
   const index=artworks.findIndex(art=>art.id===layers.current.artwork.id);
   for(let step=1;step<artworks.length;step++){
    const artwork=artworks[(index+step)%artworks.length];
    const ready=await prepareBackgroundArtwork(artwork);if(cancelled)return;
    if(ready){setLayers(previous=>({current:{game,artwork},previous:previous.current}));return;}
   }
  },18000);
  return()=>{cancelled=true;clearTimeout(timer);};
 },[game,layers.current,layers.previous,enabled,active,reduced]);
 useEffect(() => {
  if (!layers.previous) return;
  const duration=reduced?80:layers.previous.game.id===layers.current.game.id?1400:640;
  const timer = setTimeout(() => setLayers(previous => ({...previous,previous:null})),duration);
  return () => clearTimeout(timer);
 },[layers.current,layers.previous,reduced]);
 const renderLayer=(layer:ArtworkLayer,previous=false)=><div key={(previous?'previous-':'')+layerKey(layer)} className={'world-layer '+(previous?'world-previous':layers.previous?'world-enter':'')}>
  <div className="world-travel"><div className="world-drift"><Scene game={layer.game} artwork={layer.artwork}/></div></div>
 </div>;
 return <div ref={root} className="world" data-reactive={enabled&&active&&!reduced} data-drift={enabled&&!reduced} data-active={active} data-gallery={layers.previous?.game.id===layers.current.game.id} aria-hidden="true">
  <div className="world-color">
  {layers.previous&&renderLayer(layers.previous,true)}
  {renderLayer(layers.current)}
  </div>
  <div className="world-music-air"/>
  <div className="world-music-glow"/>
 </div>;
}
