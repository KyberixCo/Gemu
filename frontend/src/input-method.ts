import {useSyncExternalStore} from 'react';
import {detectControllerFamily} from './features/management/controller-devices';
import type {Family} from './features/management/controller-profiles';
export type InputMethod={kind:'keyboard'|'mouse'|'gamepad';family:Family|null;index:number|null};
let current:InputMethod={kind:'keyboard',family:null,index:null};
const listeners=new Set<()=>void>();
let stopWatching:(()=>void)|undefined;
export function inputMethod(){return current;}
export function recordInput(kind:InputMethod['kind'],id='',index:number|null=null){
 const family=kind==='gamepad'?detectControllerFamily(id):null;
 if(current.kind===kind&&current.family===family&&current.index===index)return;
 current={kind,family,index};for(const listener of listeners)listener();
}
export function reconcileInput(pads:readonly {index:number;connected:boolean}[]){if(current.kind==='gamepad'&&!pads.some(pad=>pad.connected&&pad.index===current.index))recordInput('keyboard');}
function watch(){
 const keyboard=(event:KeyboardEvent)=>{if(!['Shift','Control','Alt','Meta'].includes(event.key))recordInput('keyboard');};
 const pointer=(event:PointerEvent)=>{if(event.pointerType!=='touch')recordInput('mouse');};
 const wheel=()=>recordInput('mouse');
 let x:number|undefined,y:number|undefined;
 const move=(event:PointerEvent)=>{if(event.pointerType==='touch')return;if(x===undefined||y===undefined){x=event.clientX;y=event.clientY;return;}if(Math.hypot(event.clientX-x,event.clientY-y)>=8){x=event.clientX;y=event.clientY;recordInput('mouse');}};
 window.addEventListener('keydown',keyboard,true);window.addEventListener('pointerdown',pointer,true);window.addEventListener('wheel',wheel,{capture:true,passive:true});
 window.addEventListener('pointermove',move,{capture:true,passive:true});
 return()=>{window.removeEventListener('keydown',keyboard,true);window.removeEventListener('pointerdown',pointer,true);window.removeEventListener('wheel',wheel,true);window.removeEventListener('pointermove',move,true);};
}
function subscribe(listener:()=>void){listeners.add(listener);if(listeners.size===1)stopWatching=watch();return()=>{listeners.delete(listener);if(!listeners.size){stopWatching?.();stopWatching=undefined;}};}
export function useInputMethod(){return useSyncExternalStore(subscribe,inputMethod,inputMethod);}
export function gamepadHint(family:Family|null,action:'confirm'|'back'|'navigate'|'favorite'|'details'|'view'){
 if(action==='view')return family==='xbox'?'View':family==='ps5'?'Create':family==='ps4'?'Share':family==='switch'?'−':'Select';
 if(action==='favorite')return !family?'◧':family==='switch'?'Y':family==='xbox'?'X':'□';
 if(action==='details')return !family?'△':family==='switch'?'X':family==='xbox'?'Y':'△';
 if(action==='navigate')return '✚';
 if(!family)return action==='confirm'?'●':'○';
 return family==='switch'?(action==='confirm'?'B':'A'):family==='xbox'?(action==='confirm'?'A':'B'):(action==='confirm'?'×':'○');
}
