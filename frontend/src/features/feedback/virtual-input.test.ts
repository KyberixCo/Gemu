import {describe,expect,it} from 'vitest';
import {deleteText,editText,stickVelocity,VirtualInputHotkeys} from './virtual-input';

describe('virtual input hotkeys',()=>{
 it('opens only with Home + X, then consumes both buttons until release',()=>{
  const input=new VirtualInputHotkeys(),sample=(...buttons:number[])=>input.sample(new Set(buttons));
  sample();
  expect(sample(2).keyboard).toBe(false);sample();
  expect(sample(16)).toMatchObject({keyboard:false,suppress:true});
  expect(sample(16,2)).toMatchObject({keyboard:true,suppress:true});
  expect(sample(16,2).keyboard).toBe(false);
  expect(sample(2)).toMatchObject({keyboard:false,suppress:true});
  expect(sample()).toMatchObject({keyboard:false,suppress:false});
  expect(sample(2,16).keyboard).toBe(true);
 });
 it('consumes Home chords without triggering pointer actions or leaking X',()=>{
  const input=new VirtualInputHotkeys(),sample=(...buttons:number[])=>input.sample(new Set(buttons));
  sample();
  expect(sample(16,2)).toMatchObject({keyboard:true,pointer:false,suppress:true});
  expect(sample(2)).toMatchObject({keyboard:false,suppress:true});
  expect(sample(16,11)).toMatchObject({pointer:false,suppress:true});
  expect(sample()).toMatchObject({suppress:false});
  expect(sample(11).pointer).toBe(true);
 });
 it('does not open the keyboard or capture the View + Menu session shortcut',()=>{
  const input=new VirtualInputHotkeys();input.sample(new Set());
  expect(input.sample(new Set([8,9,2]))).toMatchObject({keyboard:false,suppress:false});
 });
 it('toggles R3 on each press without repeating or requiring neutral axes',()=>{
  const input=new VirtualInputHotkeys();input.sample(new Set());
  expect(input.sample(new Set([11])).pointer).toBe(true);
  expect(input.sample(new Set([11])).pointer).toBe(false);
  input.sample(new Set());expect(input.sample(new Set([11])).pointer).toBe(true);
 });
 it('requires release after losing focus or switching devices',()=>{
  const input=new VirtualInputHotkeys();input.sample(new Set());input.reset();
  expect(input.sample(new Set([11]))).toMatchObject({ready:false,pointer:false,suppress:true});
  input.sample(new Set());expect(input.sample(new Set([11])).pointer).toBe(true);
 });
 it('detects click, context menu and keyboard command edges before storing the next frame',()=>{
  const input=new VirtualInputHotkeys();input.sample(new Set());
  for(const button of [0,2,3,4,5]){
   expect(input.sample(new Set([button])).pressed(button)).toBe(true);
   expect(input.sample(new Set([button])).pressed(button)).toBe(false);
   input.sample(new Set());
   expect(input.sample(new Set([button])).pressed(button)).toBe(true);
   input.sample(new Set());
  }
 });
});

describe('controller text editing',()=>{
 it('inserts at the caret and replaces selected text',()=>{
  expect(editText('hola',2,2,'ñ')).toEqual({value:'hoñla',caret:3});
  expect(editText('hola',1,3,'ey')).toEqual({value:'heya',caret:3});
 });
 it('honors field limits while allowing replacement',()=>{
  expect(editText('hola',4,4,'!',4)).toEqual({value:'hola',caret:4});
  expect(editText('hola',1,3,'abc',4)).toEqual({value:'haba',caret:3});
  expect(editText('',0,0,'🎮',1)).toEqual({value:'',caret:0});
 });
 it('deletes selections and complete Unicode characters',()=>{
  expect(deleteText('hola',1,3)).toEqual({value:'ha',caret:1});
  expect(deleteText('a🎮b',3,3)).toEqual({value:'ab',caret:1});
  expect(deleteText('hola',0,0)).toEqual({value:'hola',caret:0});
 });
});
describe('virtual pointer motion',()=>{
 it('ignores drift and invalid axes',()=>{
  for(const axis of [0,0.15,-0.18,NaN,Infinity])expect(stickVelocity(axis)).toBe(0);
 });
 it('accelerates toward the stick edge and clamps malformed values',()=>{
  expect(stickVelocity(1)).toBe(1);expect(stickVelocity(-1)).toBe(-1);
  expect(stickVelocity(0.5)).toBeGreaterThan(0);expect(stickVelocity(0.5)).toBeLessThan(0.5);
  expect(stickVelocity(2)).toBe(1);
 });
});
