import {describe,it,expect,vi} from 'vitest';
import {GamepadInput,gamepadAction,pressedControls,subscribeGamepad} from './gamepad';
import {recordInput} from '../../input-method';
function pad(index=0,buttons:number[]=[],axes=[0,0],mapping:GamepadMappingType='standard'):Gamepad{
 return {index,id:`pad-${index}`,connected:true,mapping,axes,timestamp:0,vibrationActuator:{playEffect:async()=>'complete',reset:async()=>'complete'},buttons:Array.from({length:17},(_,i)=>({pressed:buttons.includes(i),value:buttons.includes(i)?1:0,touched:false}))};
}
describe('gamepad input',()=>{
 it('recognizes Home and View + Menu once until release',()=>{
  const input=new GamepadInput();expect(gamepadAction(pad(0,[16]))).toBe('menu');
  expect(input.sample([pad(0,[16])],0).action).toBe('menu');
  expect(input.sample([pad(0,[16])],1000).action).toBe('');
  input.sample([pad()],1010);
  expect(input.sample([pad(0,[8,9])],0).action).toBe('menu');
  expect(input.sample([pad(0,[8,9])],1000).action).toBe('');
  input.sample([pad()],1010);expect(input.sample([pad(0,[8,9])],1020).action).toBe('menu');
 });
 it('uses physical standard buttons and the dominant stick axis',()=>{
  expect(gamepadAction(pad(0,[0]))).toBe('confirm');expect(gamepadAction(pad(0,[1]))).toBe('back');
  expect(gamepadAction(pad(0,[2]))).toBe('favorite');expect(gamepadAction(pad(0,[3]))).toBe('details');
  expect(pressedControls(pad(0,[4,7,12]))).toEqual(['lb','rt','up']);
  expect(gamepadAction(pad(0,[],[0.8,-0.6]))).toBe('right');
  expect(gamepadAction(pad(0,[],[0.2,-0.3]))).toBe('');
  expect(gamepadAction(pad(0,[13],[1,0]))).toBe('down');
 });
 it.each([[2,'favorite'],[3,'details']] as const)('requires release before repeating button %i', (button,action)=>{
  const input=new GamepadInput();
  expect(input.sample([pad(0,[button])],0).action).toBe(action);
  expect(input.sample([pad(0,[button])],1000).action).toBe('');
  input.sample([pad()],1010);
  expect(input.sample([pad(0,[button])],1020).action).toBe(action);
 });
 it('ignores unrecognized mappings and missing axes',()=>{
  expect(gamepadAction(pad(0,[0],[],''))).toBe('');expect(gamepadAction(pad(0,[],[]))).toBe('');
 });
 it('accepts input from a second connected controller',()=>{
  const input=new GamepadInput();expect(input.sample([null,pad(),pad(1,[0])],0).action).toBe('confirm');
  expect(input.sample([pad(),pad(1,[0])],1000).action).toBe('');
 });
 it('preserves held confirmation across subscribers and releases on disconnect',()=>{
  const input=new GamepadInput();expect(input.sample([pad(0,[0])],0).action).toBe('confirm');
  expect(input.sample([pad(0,[0])],500).action).toBe('');input.sample([],510);
  expect(input.sample([pad(0,[0])],520).action).toBe('confirm');
 });
 it('requires neutral input after focus returns',()=>{
  const input=new GamepadInput();input.sample([pad(0,[0])],0,false);
  expect(input.sample([pad(0,[0])],10).action).toBe('');input.sample([pad()],20);
  expect(input.sample([pad(0,[0])],30).action).toBe('confirm');
 });
});

it('isolates virtual input and gives game menus and controller tests precedence',()=>{
 let next:FrameRequestCallback=()=>{},pads=[pad()],virtual=true,menu=false,now=0;
 vi.stubGlobal('document',Object.assign(new EventTarget(),{hidden:false,hasFocus:()=>true}));
 vi.stubGlobal('window',new EventTarget());
 vi.stubGlobal('navigator',{getGamepads:()=>pads});
 vi.stubGlobal('requestAnimationFrame',(callback:FrameRequestCallback)=>{next=callback;return 1;});
 vi.stubGlobal('cancelAnimationFrame',()=>{});
 const screen=vi.fn(),keyboard=vi.fn(),gameMenu=vi.fn();
 const stops=[subscribeGamepad({action:screen}),subscribeGamepad({capture:()=>virtual,priority:-1,action:keyboard}),subscribeGamepad({capture:()=>menu,action:gameMenu})];
 const sample=(buttons:number[])=>{pads=[pad(0,buttons)];next(now+=300);};
 try{
  sample([]);sample([0]);expect(keyboard).toHaveBeenCalledWith('confirm');expect(screen).not.toHaveBeenCalled();
  sample([0]);expect(keyboard).toHaveBeenCalledTimes(1);
  menu=true;sample([]);sample([1]);expect(gameMenu).toHaveBeenCalledWith('back');expect(keyboard).toHaveBeenCalledTimes(1);
  menu=false;virtual=false;sample([]);sample([0]);expect(screen).toHaveBeenCalledWith('confirm');
  sample([]);sample([2]);sample([]);sample([3]);expect(screen).toHaveBeenCalledTimes(1);
  const library=vi.fn();stops.push(subscribeGamepad({action:library,libraryActions:true}));
  sample([]);sample([2]);expect(library).toHaveBeenLastCalledWith('favorite');
  sample([2]);expect(library).toHaveBeenCalledTimes(1);
  sample([]);sample([3]);expect(library).toHaveBeenLastCalledWith('details');
  menu=true;sample([]);sample([2]);expect(library).toHaveBeenCalledTimes(2);
 }finally{for(const stop of stops)stop();recordInput('keyboard');vi.unstubAllGlobals();}
});
