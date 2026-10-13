import {describe,it,expect} from 'vitest';
import {setStorageProfile,profileKey,startupAction} from './profiles';
describe('Gemu profiles',()=>{
 it('requires creation, enters one profile and offers a choice for several',()=>{
  expect(startupAction([])).toBe('create');expect(startupAction([{id:'a',name:'A'}])).toBe('enter');expect(startupAction([{id:'a',name:'A'},{id:'b',name:'B'}])).toBe('select');
 });
 it('isolates every persisted setting and never reuses global keys',()=>{
  for(const key of ['gemu.favorites','gemu.management.v1','gemu.feedback']){setStorageProfile('a');const a=profileKey(key);setStorageProfile('b');expect(profileKey(key)).not.toBe(a);expect(a).not.toBe(key);}
  setStorageProfile('');expect(()=>profileKey('gemu.favorites')).toThrow();
 });
});
