import {afterEach,expect,it,vi} from 'vitest';

afterEach(()=>{vi.unstubAllGlobals();vi.resetModules();});
function storage(values:Record<string,string>={}){
 const getItem=vi.fn((key:string):string|null=>values[key]??null),setItem=vi.fn((key:string,value:string)=>{values[key]=value;});
 vi.stubGlobal('localStorage',{getItem,setItem});return {getItem,setItem,values};
}
function native(initial:Record<string,Record<string,string>>={}){
 const imported=new Set(Object.keys(initial));
 const saved=structuredClone(initial);
 const LoadPreferences=vi.fn(async(id:string,legacy:Record<string,string>|null)=>{
  saved[id]??={};
  if(!imported.has(id)&&legacy!==null){Object.assign(saved[id],legacy);imported.add(id);}
  return {values:{...saved[id]},legacyImported:imported.has(id)};
 });
 const SavePreference=vi.fn(async(id:string,key:string,value:string)=>{(saved[id]??={})[key]=value;});
 vi.stubGlobal('window',{go:{main:{App:{LoadPreferences,SavePreference}}}});
 return {LoadPreferences,SavePreference,saved};
}

it('imports only the selected player keys, retains backups, and uses SQLite on subsequent loads',async()=>{
 const browser=storage({'gemu.profile.a.gemu.favorites':'["one"]','gemu.profile.b.gemu.favorites':'["two"]','gemu.language.v1':'en'});
 const backend=native();const {hydratePreferences,preferencesStorage}=await import('./persistence');
 await hydratePreferences('a');
 expect(backend.LoadPreferences).toHaveBeenLastCalledWith('a',{'gemu.favorites':'["one"]'});
 expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["one"]');
 expect(browser.setItem).not.toHaveBeenCalled();
 browser.getItem.mockImplementation(()=>{throw new Error('browser storage unavailable');});
 await hydratePreferences('a');
 expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["one"]');
 expect(()=>preferencesStorage.getItem('gemu.profile.b.gemu.favorites')).toThrow('Preferences have not loaded');
});

it('waits for native commit and retains the previous value after a failed save',async()=>{
 storage();const backend=native({'a':{'gemu.favorites':'["old"]'}});
 const {hydratePreferences,preferencesStorage}=await import('./persistence');await hydratePreferences('a');
 let release!:()=>void;
 backend.SavePreference.mockImplementationOnce(()=>new Promise<void>(resolve=>{release=resolve;}));
 const saving=preferencesStorage.setItem('gemu.profile.a.gemu.favorites','["new"]');
 await Promise.resolve();expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["old"]');
 release();await saving;expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["new"]');
 backend.SavePreference.mockRejectedValueOnce(new Error('disk full'));
 await expect(preferencesStorage.setItem('gemu.profile.a.gemu.favorites','[]')).rejects.toThrow('disk full');
 expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["new"]');
 await preferencesStorage.setItem('gemu.profile.a.gemu.favorites','["recovered"]');
 expect(preferencesStorage.getItem('gemu.profile.a.gemu.favorites')).toBe('["recovered"]');
});

it('keeps queued writes in order and captures the owner before switching players',async()=>{
 storage();const backend=native({a:{},b:{}});const {hydratePreferences,preferencesStorage}=await import('./persistence');
 await hydratePreferences('a');await hydratePreferences('b');
 await Promise.all([preferencesStorage.setItem('gemu.profile.a.gemu.feedback','{"volume":10}'),preferencesStorage.setItem('gemu.profile.a.gemu.feedback','{"volume":20}'),preferencesStorage.setItem('gemu.profile.b.gemu.favorites','["b"]')]);
 expect(backend.SavePreference.mock.calls.map(([id,key,value])=>[id,key,value])).toEqual([['a','gemu.feedback','{"volume":10}'],['a','gemu.feedback','{"volume":20}'],['b','gemu.favorites','["b"]']]);
 expect(preferencesStorage.getItem('gemu.profile.a.gemu.feedback')).toBe('{"volume":20}');
});

it('does not complete migration when browser data is inaccessible and retries safely',async()=>{
 const browser=storage();const backend=native();browser.getItem.mockImplementation(()=>{throw new Error('blocked');});
 const {hydratePreferences}=await import('./persistence');await expect(hydratePreferences()).rejects.toThrow('blocked');
 expect(backend.LoadPreferences).toHaveBeenCalledTimes(1);
 browser.getItem.mockImplementation(key=>key==='gemu.language.v1'?'en':null);await hydratePreferences();
 expect(backend.saved['']['gemu.language.v1']).toBe('en');
});

it('keeps browser previews local and requires an updated native host',async()=>{
 const browser=storage();const {preferencesStorage,hydratePreferences}=await import('./persistence');
 await preferencesStorage.setItem('gemu.language.v1','es');expect(browser.values['gemu.language.v1']).toBe('es');
 vi.stubGlobal('window',{go:{main:{App:{}}}});await expect(hydratePreferences()).rejects.toThrow('Reinicia Gemu');
});
