import {describe,it,expect} from 'vitest';
import {defaultProfile,label,validProfiles} from './controller-profiles';
describe('controller profile contracts',()=>{
 it('keeps Nintendo face labels in their physical positions',()=>{
  expect(['south','east','west','north'].map(id=>label('switch',id as 'south'))).toEqual(['B','A','Y','X']);
  expect(label('xbox','south')).toBe('A');
  expect(label('ps4','south')).toBe('×');
  expect(label('ps5','select')).toBe('Crear');
 });
 it('creates independent profiles for emulator overrides',()=>{
  const global=defaultProfile(),retroarch=defaultProfile();
  retroarch.bindings.select='Menú del emulador';
  expect(global.bindings.select).toBe('Select');
  expect(validProfiles({'Global:xbox':global,'RetroArch:xbox':retroarch})).toBe(true);
 });
 it('rejects corrupted persisted values and unsupported destinations',()=>{
  expect(validProfiles({'RetroArch:switch':{...defaultProfile(),deadzone:Infinity}})).toBe(false);
  expect(validProfiles({'RetroArch:ps5':{...defaultProfile(),bindings:{south:'execute'}}})).toBe(false);
  expect(validProfiles({'unknown:xbox':defaultProfile()})).toBe(false);
  expect(validProfiles(null)).toBe(false);
 });
});
