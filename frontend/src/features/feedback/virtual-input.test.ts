import {describe,expect,it} from 'vitest';
import {deleteText,editText,stickVelocity} from './virtual-input';

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
