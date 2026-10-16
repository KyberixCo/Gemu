import {GamepadRepeater} from './input';
import type {Control} from '../management/controller-profiles';
import {browserController} from '../controllers/providers';
import {controllerAssociations} from '../controllers/registry';
import {isForeground,subscribeActivity} from '../../activity';
import {recordInput,reconcileInput} from '../../input-method';

export type GamepadAction='confirm'|'back'|'menu'|'favorite'|'details'|'up'|'down'|'left'|'right'|'';
export const buttonControls:readonly Control[]=['south','east','west','north','lb','rb','lt','rt','select','start','ls','rs','up','down','left','right'];
export function pressedControls(pad:Gamepad):Control[]{
 return pad.mapping==='standard'?buttonControls.filter((_,index)=>pad.buttons[index]?.pressed):[];
}
export function gamepadAction(pad:Gamepad|null):GamepadAction{
 if(!pad||pad.mapping!=='standard')return '';
 if(pad.buttons[16]?.pressed)return 'menu';
 if(pad.buttons[8]?.pressed&&pad.buttons[9]?.pressed)return 'menu';
 const pressed=pressedControls(pad);
 if(pressed.includes('south'))return 'confirm';
 if(pressed.includes('east'))return 'back';
 if(pressed.includes('west'))return 'favorite';
 if(pressed.includes('north'))return 'details';
 for(const direction of ['up','down','left','right'] as const)if(pressed.includes(direction))return direction;
 const x=Number.isFinite(pad.axes[0])?pad.axes[0]:0,y=Number.isFinite(pad.axes[1])?pad.axes[1]:0;
 if(Math.max(Math.abs(x),Math.abs(y))<0.5)return '';
 return Math.abs(x)>Math.abs(y)?x<0?'left':'right':y<0?'up':'down';
}
export class GamepadInput{
 private index:number|null=null;
 private repeater=new GamepadRepeater();
 private waitForRelease=false;
 sample(pads:readonly (Gamepad|null)[],now:number,active=true){
  const supported=pads.filter((pad):pad is Gamepad=>!!pad?.connected&&pad.mapping==='standard');
  const pad=supported.find(pad=>gamepadAction(pad))??supported.find(pad=>pad.index===this.index)??supported[0]??null;
  if(pad?.index!==this.index){this.repeater.reset();this.index=pad?.index??null;}
  const next=gamepadAction(pad);
  if(!active){this.waitForRelease=true;this.repeater.reset();return {pad,action:'' as GamepadAction};}
  if(this.waitForRelease){if(!next)this.waitForRelease=false;return {pad,action:'' as GamepadAction};}
  return {pad,action:this.repeater.accept(next,now)?next:'' as GamepadAction};
 }
}
type Listener={action?:(action:GamepadAction)=>void;libraryActions?:boolean;snapshot?:(pad:Gamepad|null)=>void;devices?:(pads:readonly Gamepad[],active:Gamepad|null)=>void;capture?:()=>boolean;priority?:number};
const listeners:Listener[]=[];
const input=new GamepadInput();
let frame=0;
const physical=new Map<number,string>();
let stopWatching:(()=>void)|undefined;
function syncPolling(){
 cancelAnimationFrame(frame);frame=0;
 if(listeners.length&&isForeground())frame=requestAnimationFrame(poll);
 else input.sample([],performance.now(),false);
}
function watchActivity(){
 const unsubscribe=subscribeActivity(syncPolling);
 window.addEventListener('focus',syncPolling);window.addEventListener('blur',syncPolling);document.addEventListener('visibilitychange',syncPolling);
 return()=>{unsubscribe();window.removeEventListener('focus',syncPolling);window.removeEventListener('blur',syncPolling);document.removeEventListener('visibilitychange',syncPolling);};
}
function poll(now:number){
 let pads:readonly (Gamepad|null)[]=[];
 try{pads=navigator.getGamepads?.()??[];}catch{/* API may be unavailable in the host. */}
 controllerAssociations.observe(pads.filter((pad):pad is Gamepad=>!!pad?.connected).map(pad=>browserController(pad.index,pad.id)));
 const active=isForeground();
 const {pad,action}=input.sample(pads,now,active);
 if(active){
  const connected=pads.filter((device):device is Gamepad=>!!device?.connected);
  reconcileInput(connected);
  for(const device of connected){
   if(device.mapping!=='standard')continue;
   const pressed=pressedControls(device).join(',')+(device.buttons[16]?.pressed?',guide':'');
   if(pressed&&physical.get(device.index)!==pressed)recordInput('gamepad',device.id,device.index);
   physical.set(device.index,pressed);
  }
  for(const index of physical.keys())if(!connected.some(device=>device.index===index))physical.delete(index);
  if(action&&pad)recordInput('gamepad',pad.id,pad.index);
 }
 for(const listener of [...listeners]){
  listener.snapshot?.(active?pad:null);
  listener.devices?.(pads.filter((device):device is Gamepad=>!!device?.connected),active?pad:null);
 }
 const capture=listeners.filter(listener=>listener.capture?.()).sort((a,b)=>(b.priority??0)-(a.priority??0))[0];
 const recipient=capture??listeners.filter(listener=>listener.action&&!listener.capture).at(-1);
 if(action&&(action!=='menu'||capture)&&(!['favorite','details'].includes(action)||recipient?.libraryActions))recipient?.action?.(action);
 if(listeners.length&&isForeground())frame=requestAnimationFrame(poll);
}
export function subscribeGamepad(listener:Listener){
 listeners.push(listener);
 if(listeners.length===1){stopWatching=watchActivity();syncPolling();}
 return ()=>{const index=listeners.indexOf(listener);if(index>=0)listeners.splice(index,1);if(!listeners.length){cancelAnimationFrame(frame);stopWatching?.();stopWatching=undefined;}};
}
export function hasGamepadCapture(except:()=>boolean){return listeners.some(listener=>listener.capture!==except&&listener.capture?.());}
