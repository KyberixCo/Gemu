import {describe,it,expect} from 'vitest';
import {GamepadRepeater} from './input';
describe('controller repetition',()=>{
 it('keeps confirmation held across screen changes until release',()=>{
  const input=new GamepadRepeater();expect(input.accept('confirm',0)).toBe(true);
  expect(input.accept('confirm',16)).toBe(false);expect(input.accept('confirm',1900)).toBe(false);
  expect(input.accept('',1910)).toBe(false);expect(input.accept('confirm',1920)).toBe(true);
 });
 it('repeats directions at a bounded rate',()=>{
  const input=new GamepadRepeater();expect(input.accept('right',0)).toBe(true);expect(input.accept('right',200)).toBe(false);
  expect(input.accept('right',241)).toBe(true);expect(input.accept('left',242)).toBe(true);
 });
 it('does not repeat back, and resets after losing focus',()=>{
  const input=new GamepadRepeater();expect(input.accept('back',0)).toBe(true);expect(input.accept('back',500)).toBe(false);
  input.reset();expect(input.accept('back',510)).toBe(true);
 });
});
