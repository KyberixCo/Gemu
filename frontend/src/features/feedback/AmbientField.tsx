import { useEffect, useRef } from 'react';
import { interfaceAudio } from './audio';
import { ambientMotion } from './ambient-motion';

// Light within the scene, rather than a separate set of floating UI objects.
export function AmbientField({ enabled, active, reduced, welcome=false }: {enabled:boolean;active:boolean;reduced:boolean;welcome?:boolean}) {
 const root=useRef<HTMLDivElement>(null);
 const visualSeconds=useRef(0);
 useEffect(()=>{
  if(!enabled||!active||reduced)return;
  const element=root.current;if(!element)return;
  const veils=[...element.querySelectorAll<HTMLElement>('.ambient-veil')];
  const reflection=element.querySelector<HTMLElement>('.ambient-reflection');
  let frame=0,last=0,smoothed=0;
  const draw=(now:number)=>{
   if(now-last>=50){
    const delta=last?Math.min((now-last)/1000,.1):0;last=now;
    const audio=interfaceAudio.ambientFrame();
    if(audio.playing)visualSeconds.current=audio.seconds;else visualSeconds.current+=delta;
    smoothed+=(audio.level-smoothed)*.18;
    const motion=ambientMotion(visualSeconds.current,smoothed,audio.pulse);
    const cycle=motion.phase*Math.PI/12;
    veils.forEach((veil,i)=>{
     const drift=Math.sin(cycle+i*1.4),depth=Math.cos(cycle+i*.8);
     veil.style.transform=`translate3d(${drift*(12+i*8)}px,${depth*(8+i*4)}px,0) scale(${1.035+motion.breath*.018})`;
     veil.style.opacity=String(.12+motion.intensity*(.28-i*.045));
    });
    if(reflection){
     reflection.style.transform=`translate3d(${Math.sin(cycle)*55}px,${Math.cos(cycle)*12}px,0)`;
     reflection.style.opacity=String(.03+motion.pulse*.13+smoothed*.05);
    }
   }
   frame=requestAnimationFrame(draw);
  };frame=requestAnimationFrame(draw);return()=>cancelAnimationFrame(frame);
 },[enabled,active,reduced]);
 if(!enabled)return null;
 return <div ref={root} className={'ambient-field '+(welcome?'ambient-welcome':'')} data-paused={!active||reduced} aria-hidden="true">
  <div className="ambient-veil ambient-depth"/>
  <div className="ambient-veil ambient-air"/>
  <div className="ambient-veil ambient-afterglow"/>
  <div className="ambient-reflection"/>
 </div>;
}
