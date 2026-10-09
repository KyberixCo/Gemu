import { describe,it,expect } from 'vitest';
import { defaultPreferences,isReducedMotion,parsePreferences } from './preferences';
describe('feedback preferences',()=>{
 it('recovers defaults from invalid or missing storage',()=>{expect(parsePreferences('{')).toEqual(defaultPreferences);expect(parsePreferences(null)).toEqual(defaultPreferences);});
 it('clamps volume and rejects invalid types',()=>{
  expect(parsePreferences('{"volume":150,"motion":"unknown","sound":"false"}')).toEqual({...defaultPreferences,volume:100});
  expect(parsePreferences('{"volume":-1,"welcomeSound":false}')).toEqual({...defaultPreferences,volume:0,welcomeSound:false});
 });
 it('migrates existing preferences and validates the independent ambient controls',()=>{
  expect(parsePreferences('{"sound":false}')).toEqual({...defaultPreferences,sound:false});
  expect(parsePreferences('{"ambient":false,"ambientVolume":110,"ambientMotion":false}')).toEqual({...defaultPreferences,ambient:false,ambientVolume:100,ambientMotion:false});
 });
 it('uses system motion unless explicitly overridden',()=>{
  expect(isReducedMotion('system',true)).toBe(true);expect(isReducedMotion('system',false)).toBe(false);
  expect(isReducedMotion('full',true)).toBe(false);expect(isReducedMotion('reduced',false)).toBe(true);
 });
});
