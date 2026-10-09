import {afterEach,beforeEach,describe,expect,it,vi} from 'vitest';
import {FakeContext} from '../audio-test-support';
import {ProceduralMusic} from './engine';
import {profileForGame} from './profile';
import {BEAT_SECONDS} from './composition';
import {games} from '../../../catalog';
let context:FakeContext;
function create(){return new ProceduralMusic(context as unknown as AudioContext,context.destination as AudioNode,profileForGame(games[0]));}
beforeEach(()=>{vi.useFakeTimers();context=new FakeContext();});
afterEach(()=>{vi.clearAllTimers();vi.useRealTimers();});
describe('musical transport',()=>{
 it('uses emitted textures and sustains for light without anticipating future notes',()=>{
  const music=create();music.start();
  expect(music.frame().warmth).toBe(0);expect(music.frame().crystal).toBe(0);
  context.currentTime=5.7;expect(Math.abs(music.frame().warmth)).toBeGreaterThan(0);
  // Release mocked voices so texture scheduling represents an audible ensemble.
  for(const source of [...context.sources])source.onended?.();
  context.currentTime=5+BEAT_SECONDS*10;music.tick();expect(music.frame().wind).toBe(0);
  context.currentTime+=2.5;expect(music.frame().wind).toBeGreaterThan(0);
  context.currentTime=5+BEAT_SECONDS*12;music.tick();expect(music.frame().crystal).toBe(0);
  context.currentTime+=.5;expect(music.frame().crystal).toBeGreaterThan(0);
  music.stop();expect(music.frame()).toMatchObject({pulse:0,crystal:0,wind:0,warmth:0});
 });
 it('preserves notes and the shared room across queued changes',()=>{
  const music=create();music.start();const sources=[...context.sources],filters=context.filters.length;
  expect(sources.length).toBeGreaterThan(0);
  const scheduledStops=sources.map(source=>source.stop.mock.calls.map(call=>[...call]));
  context.currentTime+=BEAT_SECONDS;music.select(profileForGame(games[1]));music.select(profileForGame(games[2]));music.tick();
  expect(sources.map(source=>source.stop.mock.calls)).toEqual(scheduledStops);expect(context.filters.length).toBe(filters);
  context.currentTime=5+BEAT_SECONDS*4;music.tick();
  expect(context.filters[0].frequency.setTargetAtTime).toHaveBeenCalledWith(profileForGame(games[2]).brightness,expect.any(Number),2.4);
 });
 it('skips overdue beats without scheduling a catch-up burst',()=>{
  const music=create();music.start();context.currentTime+=120;const count=context.sources.length;music.tick();
  expect(context.sources.length-count).toBeLessThanOrEqual(5);
  expect(context.sources.slice(count).every(s=>s.start.mock.calls[0][0]>=context.currentTime)).toBe(true);
 });
 it('stops the scheduler and future notes when muted or inactive',()=>{
  const music=create();music.start();music.stop();expect(vi.getTimerCount()).toBe(0);expect(music.frame().playing).toBe(false);
  expect(context.sources.every(s=>s.stop.mock.calls.at(-1)![0]===5.025)).toBe(true);
 });
 it('bounds the ensemble without stealing musical tails',()=>{
  const music=create();music.start();
  // Mocked nodes never finish: even under this worst case the cap remains fixed.
  for(let beat=1;beat<600;beat++){context.currentTime=5+beat*BEAT_SECONDS;music.tick();}
  expect(context.sources.length).toBeLessThanOrEqual(24);expect(context.sources.every(s=>s.stop.mock.calls.length===1)).toBe(true);
 });
 it('drives light cues from emitted notes rather than a fixed cycle',()=>{
  const music=create();music.start();let found=false;
  for(let beat=1;beat<64;beat++){
   context.currentTime=5+beat*BEAT_SECONDS;music.tick();
   context.currentTime+=.5;
   if(music.frame().pulse>0){found=true;break;}
  }
  expect(found).toBe(true);music.stop();expect(music.frame().pulse).toBe(0);
 });
});
