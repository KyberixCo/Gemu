import {afterEach,expect,it,vi} from 'vitest';
import {chooseCollections,installSize,isInstalling,retroHost,type RetroJob} from './retroarch';
import {loadManagementConfig,defaults} from './management-config';
import {setStorageProfile} from '../../profiles';
import {translate} from '../../localization';
afterEach(()=>vi.unstubAllGlobals());
it('retains configured console folders when systems change and creates only new defaults',()=>{
 const before=[{systemId:'nes',directory:'D:/My ROMs',recursive:false},{systemId:'snes',directory:'E:/SNES',recursive:true}];
 expect(chooseCollections(before,['snes','gb'])).toEqual([before[1],{systemId:'gb',directory:'',recursive:true}]);expect(before).toHaveLength(2);
});
it('distinguishes progress, terminal states and unknown download sizes',()=>{
 for(const state of ['starting','downloading','extracting','publishing'])expect(isInstalling({state} as RetroJob)).toBe(true);
 for(const state of ['ready','cancelled','error'])expect(isInstalling({state} as RetroJob)).toBe(false);
 expect(isInstalling(undefined)).toBe(false);
 expect(installSize({token:'t',version:'1',destination:'d',downloads:[{name:'a',url:'u',size:100},{name:'b',url:'u',size:200}]})).toBe(300);
 expect(installSize({token:'t',version:'1',destination:'d',downloads:[{name:'a',url:'u',size:-1}]})).toBeNull();
});
it('upgrades old management preferences without resetting custom sources or emulator paths',()=>{
 const saved=structuredClone(defaults);delete saved.sources.RetroArch;saved.sources.Steam.path='D:/Steam';saved.emulators.RetroArch.binary='D:/RetroArch/retroarch.exe';
 vi.stubGlobal('localStorage',{getItem:()=>JSON.stringify(saved)});setStorageProfile('player');
 const result=loadManagementConfig();expect(result.sources.RetroArch).toEqual({enabled:false,path:''});expect(result.sources.Steam.path).toBe('D:/Steam');expect(result.emulators.RetroArch.binary).toBe('D:/RetroArch/retroarch.exe');
});
it('requires the new host instead of exposing a nonfunctional install action',()=>{
 vi.stubGlobal('window',{go:{main:{App:{GetSources:async()=>[]}}}});expect(()=>retroHost()).toThrow('Reinicia Gemu');
});
it('recovers a missing or malformed RetroArch source without dropping valid older preferences',()=>{
 const saved=structuredClone(defaults);saved.sources.Steam.path='D:/Keep';
 (saved.sources as Record<string,unknown>).RetroArch=null;
 vi.stubGlobal('localStorage',{getItem:()=>JSON.stringify(saved)});setStorageProfile('player');
 expect(loadManagementConfig().sources.Steam.path).toBe('D:/Keep');expect(loadManagementConfig().sources.RetroArch).toEqual({enabled:false,path:''});
});
it('translates setup actions and diagnostics in both languages',()=>{
 expect(translate('Instalar y configurar RetroArch','en')).toBe('Install and configure RetroArch');
 expect(translate('Reinicia Gemu para instalar y configurar RetroArch.','en')).toContain('Restart Gemu');
 expect(translate('RetroArch archive integrity check failed','es')).toContain('integridad');
 expect(translate('Usar esta instalación','es')).toBe('Usar esta instalación');
});
