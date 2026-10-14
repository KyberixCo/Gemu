import {afterEach,expect,it,vi} from 'vitest';
import {assignKey,keyboardDefaults,navigationKey,readKeyboardBindings,validKeyboardBindings} from './keyboard-bindings';
import {setStorageProfile,profileKey} from './profiles';
afterEach(()=>{setStorageProfile('');vi.unstubAllGlobals();});
function event(key:string,code:string,target:object|null=null){return {key,code,target,preventDefault:vi.fn()} as unknown as KeyboardEvent;}
it('swaps conflicts and rejects incomplete or invalid mappings',()=>{
 const swapped=assignKey(keyboardDefaults,'confirm','Escape');expect(swapped.confirm).toBe('Escape');expect(swapped.back).toBe('Enter');expect(validKeyboardBindings(swapped)).toBe(true);
 expect(validKeyboardBindings({...keyboardDefaults,back:'Enter'})).toBe(false);expect(validKeyboardBindings({})).toBe(false);expect(assignKey(keyboardDefaults,'confirm','F10')).toEqual(keyboardDefaults);
});
it('uses saved physical keys without intercepting text entry or shortcuts',()=>{
 setStorageProfile('player');vi.stubGlobal('localStorage',{getItem:(key:string)=>key===profileKey('keyboard.v1')?JSON.stringify({...keyboardDefaults,confirm:'KeyE'}):null});
 expect(navigationKey(event('e','KeyE'))).toBe('Enter');
 const old=event('Enter','Enter');expect(navigationKey(old)).toBe('');expect(old.preventDefault).toHaveBeenCalled();
 expect(navigationKey(event('e','KeyE',{tagName:'INPUT'}))).toBe('e');
 expect(navigationKey({...event('e','KeyE'),ctrlKey:true} as KeyboardEvent)).toBe('');
 expect(navigationKey(event('Tab','Tab'))).toBe('Tab');expect(navigationKey(event('F10','F10'))).toBe('F10');
});
it('isolates profiles and recovers from malformed saved data',()=>{
 setStorageProfile('one');vi.stubGlobal('localStorage',{getItem:(key:string)=>key.includes('.one.')?JSON.stringify({...keyboardDefaults,confirm:'KeyE'}):'{broken'});
 expect(readKeyboardBindings().confirm).toBe('KeyE');setStorageProfile('two');expect(readKeyboardBindings()).toEqual(keyboardDefaults);
});
