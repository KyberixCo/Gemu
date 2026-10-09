import {describe,it,expect} from 'vitest';
import {ambientMotion} from './ambient-motion';
import {BEAT_SECONDS} from './music/composition';
describe('ambient light clock',()=>{
 it('breathes with the musical clock while keeping scene drift bounded',()=>{
  expect(ambientMotion(BEAT_SECONDS*2,0).breath).toBeCloseTo(1);
  expect(ambientMotion(BEAT_SECONDS*10,0).breath).toBeCloseTo(1);
  expect(ambientMotion(27,0).phase).toBe(3);
 });
 it('responds only to actual musical note cues',()=>{
  expect(ambientMotion(3,0).pulse).toBe(0);expect(ambientMotion(3,0,.7).pulse).toBe(.7);
 });
 it('uses waveform energy without inventing notes or exceeding bounds',()=>{
  const silent=ambientMotion(7,0,.4),loud=ambientMotion(7,1,.4);
  expect(loud.pulse).toEqual(silent.pulse);expect(loud.intensity).toBeGreaterThan(silent.intensity);
  expect(ambientMotion(3.1,100,100).intensity).toBeLessThanOrEqual(1);
 });
});
