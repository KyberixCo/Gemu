import {describe,it,expect} from 'vitest';
import {FakeContext} from '../audio-test-support';
import {instrumentBuffers,roomImpulse} from './synthesis';
describe('original instrument synthesis',()=>{
 it('creates finite, bounded samples and a periodic pad seam',()=>{
  const context=new FakeContext();context.sampleRate=48000;const buffers=instrumentBuffers(context as unknown as AudioContext);
  for(const buffer of Object.values(buffers))for(let channel=0;channel<2;channel++){
   const values=buffer.getChannelData(channel);expect(values.every(Number.isFinite)).toBe(true);expect(values.every(value=>Math.abs(value)<.5)).toBe(true);
  }
  const pad=buffers.pad.getChannelData(0),seam=pad[0]-pad.at(-1)!,next=pad[1]-pad[0];
  expect(Math.abs(seam-next)).toBeLessThan(.001);
  expect(buffers.bell.getChannelData(0).at(-1)).toBe(0);expect(buffers.pluck.getChannelData(0).at(-1)).toBe(0);
  for(const instrument of ['wind','gust','surf','rustle','crystal'] as const){
   const left=buffers[instrument].getChannelData(0),right=buffers[instrument].getChannelData(1);
   expect(left[0]).toBeCloseTo(0);expect(left.at(-1)).toBeCloseTo(0);
   expect(left.some((sample,i)=>sample!==right[i])).toBe(true);
   expect(left.some(sample=>Math.abs(sample)>.001)).toBe(true);
  }
  expect(buffers.wind.duration).toBe(8);
 });
 it('creates a stereo room with a quiet ending',()=>{
  const context=new FakeContext(),impulse=roomImpulse(context as unknown as AudioContext);
  const left=impulse.getChannelData(0),right=impulse.getChannelData(1);
  expect(left.some((sample,i)=>sample!==right[i])).toBe(true);expect(left.at(-1)).toBeCloseTo(0);expect(right.at(-1)).toBeCloseTo(0);
 });
});
