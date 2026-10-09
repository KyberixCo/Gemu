import {describe,expect,it} from 'vitest';
import {sceneLighting} from './scene-lighting';
const silence={playing:true,level:0,pulse:0,crystal:0,wind:0,warmth:0,pan:0};
describe('music-driven cover lighting',()=>{
 it('restores the original artwork when audio is stopped',()=>{
  expect(sceneLighting({...silence,playing:false,crystal:1,wind:1,warmth:1})).toEqual(sceneLighting(silence));
 });
 it('separates crystal reflections, wind haze and harmonic tint',()=>{
  const crystal=sceneLighting({...silence,crystal:1,pan:-.5}),wind=sceneLighting({...silence,wind:1});
  expect(crystal.glow).toBeGreaterThan(0);expect(crystal.air).toBe(0);expect(crystal.position).toBeLessThan(74);
  expect(wind.air).toBeGreaterThan(0);expect(wind.glow).toBe(0);
  expect(sceneLighting({...silence,warmth:-1}).hue).toBe(-28);
  expect(sceneLighting({...silence,warmth:1}).hue).toBe(28);
 });
 it('bounds color and brightness even with excessive or invalid inputs',()=>{
  const light=sceneLighting({...silence,level:100,pulse:100,crystal:100,wind:100,warmth:100,pan:100});
  expect(light.brightness).toBeLessThan(1.64);expect(light.saturation).toBeLessThan(1.87);
  expect(light.glow).toBeLessThan(.75);expect(light.position).toBe(92);
  expect(Object.values(sceneLighting({...silence,level:NaN,wind:Infinity})).every(Number.isFinite)).toBe(true);
 });
 it('makes quiet musical cues visible without lighting silent passages',()=>{
  const light=sceneLighting({...silence,level:.04,pulse:.1,crystal:.1,wind:.25,warmth:.16});
  expect(light.brightness).toBeGreaterThan(1.15);expect(light.glow).toBeGreaterThan(.2);
  expect(light.air).toBeCloseTo(.15);expect(light.hue).toBeGreaterThan(10);
  expect(sceneLighting(silence)).toMatchObject({brightness:1,saturation:1,hue:0,glow:0,air:0});
 });
});
