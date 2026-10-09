import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest';
import { InterfaceAudio } from './audio';
import { defaultPreferences } from './preferences';
import { SoundPolicy } from './policy';

import {FakeContext} from './audio-test-support';
import {games} from '../../catalog';
beforeEach(()=>{
 vi.useFakeTimers();FakeContext.instances=[];
 vi.stubGlobal('AudioContext',FakeContext);
 vi.stubGlobal('fetch',vi.fn(async()=>({ok:true,arrayBuffer:async()=>new ArrayBuffer(8)})));
});
afterEach(()=>{vi.clearAllTimers();vi.useRealTimers();vi.unstubAllGlobals();});
describe('audio lifecycle',()=>{
 it('loads and decodes once, sharing one context',async()=>{
  const audio=new InterfaceAudio();await Promise.all([audio.unlock(),audio.unlock()]);await audio.unlock();
  expect(FakeContext.instances).toHaveLength(1);expect(fetch).toHaveBeenCalledTimes(9);expect(FakeContext.instances[0].decodeAudioData).toHaveBeenCalledTimes(9);
 });
 it('does not queue navigation before activation',async()=>{
  const audio=new InterfaceAudio();expect(audio.play('navigate')).toBe(false);await audio.unlock();
  expect(FakeContext.instances[0].sources).toHaveLength(0);
 });
 it('silences active voices when muted or volume is zero',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();expect(audio.play('open')).toBe(true);
  audio.configure({...defaultPreferences,sound:false});expect(audio.play('close')).toBe(false);
  expect(FakeContext.instances[0].sources[0].stop).toHaveBeenCalledWith(5.025);
  audio.configure({...defaultPreferences,volume:0});expect(audio.play('confirm')).toBe(false);
 });
 it('drops a welcome awaiting unlock when skipped',async()=>{
  const audio=new InterfaceAudio();const pending=audio.preview('welcome');audio.stop();expect(await pending).toBe(false);
  expect(FakeContext.instances[0].sources).toHaveLength(0);
 });
 it('lets the baked reverb finish when welcome visuals complete naturally',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.play('welcome');audio.cancelPending();
  expect(FakeContext.instances[0].sources[0].stop).not.toHaveBeenCalled();
  audio.stop();expect(FakeContext.instances[0].sources[0].stop).toHaveBeenCalled();
 });
 it('invalidates delayed welcome without cutting an already active tail',async()=>{
  const audio=new InterfaceAudio();const pending=audio.preview('welcome');audio.cancelPending();
  expect(await pending).toBe(false);expect(FakeContext.instances[0].sources).toHaveLength(0);
 });
 it('drops effects while hidden, and never replays on return',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.play('open');audio.setActive(false);
  expect(audio.play('confirm')).toBe(false);audio.setActive(true);expect(FakeContext.instances[0].sources).toHaveLength(1);
 });
 it('honors the independent welcome sound preference',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.configure({...defaultPreferences,welcomeSound:false});
  expect(audio.play('welcome')).toBe(false);expect(audio.play('confirm')).toBe(true);
 });
 it('does not allow a navigation tick to evict important sounds',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.play('welcome');audio.play('error');
  expect(audio.play('navigate')).toBe(false);expect(FakeContext.instances[0].sources).toHaveLength(2);
 });
 it('releases stopped navigation when confirming',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.play('navigate');audio.play('open');
  expect(FakeContext.instances[0].sources[0].stop).toHaveBeenCalled();
 });
 it('keeps the UI silent if files are unavailable',async()=>{
  vi.stubGlobal('fetch',vi.fn(async()=>{throw new Error('offline');}));
  const audio=new InterfaceAudio();expect(await audio.unlock()).toBe(true);expect(audio.play('open')).toBe(false);
 });
 it('tolerates unsupported audio and rejected activation',async()=>{
  vi.stubGlobal('AudioContext',class {constructor(){throw new Error('not supported');}});
  expect(await new InterfaceAudio().preview('welcome')).toBe(false);
 });
});
describe('navigation admission',()=>{
 it('does not tick on unchanged focus or while disabled',()=>{
  const policy=new SoundPolicy();expect(policy.accept('navigate',0,true,false)).toBe(false);expect(policy.accept('navigate',0,false)).toBe(false);
 });
 it('limits repeated navigation to one tick per 90ms',()=>{
  const policy=new SoundPolicy();expect(policy.accept('navigate',0,true)).toBe(true);expect(policy.accept('navigate',40,true)).toBe(false);
  expect(policy.accept('navigate',90,true)).toBe(true);policy.reset();expect(policy.accept('navigate',91,true)).toBe(true);
 });
});

describe('procedural music bus',()=>{
 it('starts one transport only after entering the experience',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();expect(FakeContext.instances[0].sources).toHaveLength(0);
  audio.setAmbientSession(true);const count=FakeContext.instances[0].sources.length;
  expect(count).toBeGreaterThanOrEqual(3);audio.setAmbientSession(true);await audio.unlock();
  expect(FakeContext.instances[0].sources).toHaveLength(count);expect(vi.getTimerCount()).toBe(1);
 });
 it('keeps music volume independent of interface sounds',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  audio.configure({...defaultPreferences,sound:false,volume:0});
  expect(audio.ambientFrame().playing).toBe(true);expect(audio.play('navigate')).toBe(false);
  expect(FakeContext.instances[0].sources[0].stop).toHaveBeenCalledTimes(1); // Natural note end only.
 });
 it('turns off the background without muting confirmations',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  audio.configure({...defaultPreferences,ambient:false});
  expect(audio.ambientFrame().playing).toBe(false);expect(audio.play('confirm')).toBe(true);expect(vi.getTimerCount()).toBe(0);
 });
 it('pauses on focus loss and resumes at the same musical position',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  const context=FakeContext.instances[0];context.currentTime=6.25;audio.setActive(false);
  expect(audio.ambientFrame().playing).toBe(false);context.currentTime=12;audio.setActive(true);
  expect(audio.ambientFrame().seconds).toBeCloseTo(1.25);expect(vi.getTimerCount()).toBe(1);
 });
 it('ducks music under confirmations and isolates it from UI voices',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);audio.play('confirm');
  const context=FakeContext.instances[0];expect(context.gains[2].gain.setTargetAtTime).toHaveBeenCalledWith(.205,5,.18);
  expect(context.gains[2].connect).toHaveBeenCalledWith(context.analysers[0]);
  expect(context.gains.at(-1)!.connect).toHaveBeenCalledWith(context.gains[0]);expect(audio.ambientFrame().level).toBeCloseTo(.2);
 });
 it('keeps music level steady during navigation and panel changes',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  const output=FakeContext.instances[0].gains[2].gain.setTargetAtTime;
  output.mockClear();
  for(const name of ['navigate','section','open','close'] as const) expect(audio.play(name)).toBe(true);
  expect(output).not.toHaveBeenCalled();
 });
 it('cancels a pending start when the session becomes ineligible',async()=>{
  const audio=new InterfaceAudio();const unlocked=audio.unlock();audio.setAmbientSession(true);audio.setAmbientSession(false);
  await unlocked;expect(FakeContext.instances[0].sources).toHaveLength(0);
 });
 it('silences music at volume zero and resumes without duplicate transports',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  audio.configure({...defaultPreferences,ambientVolume:0});expect(audio.ambientFrame().playing).toBe(false);expect(vi.getTimerCount()).toBe(0);
  audio.configure(defaultPreferences);audio.configure(defaultPreferences);expect(vi.getTimerCount()).toBe(1);expect(audio.ambientFrame().playing).toBe(true);
 });
 it('requires no ambient recording or network service',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);
  expect(fetch).toHaveBeenCalledTimes(9);expect(audio.ambientFrame().playing).toBe(true);
 });
 it('changes selected games without stopping notes or resetting time',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();audio.setAmbientSession(true);const context=FakeContext.instances[0];
  context.currentTime=6.25;const sources=[...context.sources],stops=sources.map(source=>source.stop.mock.calls.length);
  audio.selectGame(games[1]);audio.selectGame(games[2]);
  expect(audio.ambientFrame().seconds).toBeCloseTo(1.25);expect(sources.map(source=>source.stop.mock.calls.length)).toEqual(stops);
 });
 it('keeps UI effects available if synthesis is unsupported',async()=>{
  const audio=new InterfaceAudio();await audio.unlock();const context=FakeContext.instances[0];
  vi.spyOn(context,'createConvolver').mockImplementation(()=>{throw new Error('unsupported');});audio.setAmbientSession(true);
  expect(audio.ambientFrame().playing).toBe(false);expect(audio.play('open')).toBe(true);
 });
});
