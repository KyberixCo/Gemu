import {expect,it} from 'vitest';
import {providerActivation,providerReadiness,providerSteps,sameCredentialValues} from './setup-providers';
it('asks only for the personal account for ScreenScraper',()=>{
 expect(providerSteps('screenscraper',true)).toEqual(['Introducción','Tu cuenta','Comprobación','Preferencias','Revisión']);
 expect(providerSteps('steamgriddb',false)).toEqual(['Introducción','Credenciales','Comprobación','Revisión']);
});
it('allows anonymous checks and requires a complete personal account',()=>{
 const state=providerReadiness('screenscraper',{username:'user',password:'password'},[]);
 expect(state.ready).toBe(true);expect(state.missing).toEqual([]);
 expect(providerReadiness('screenscraper',{},[]).ready).toBe(true);
 expect(providerReadiness('screenscraper',{username:'user'},[]).incompleteAccount).toBe(true);
 expect(providerReadiness('screenscraper',{username:'user'},['password']).ready).toBe(true);
});
it('keeps a successful check across equivalent parent rerenders but invalidates changed secrets',()=>{
 expect(sameCredentialValues({apiKey:'key'},{apiKey:'key'})).toBe(true);
 expect(sameCredentialValues({}, {apiKey:''})).toBe(true);
 expect(sameCredentialValues({apiKey:'key'},{apiKey:'new'})).toBe(false);
});
it('activates checked integrations but keeps unverified or unimplemented connections pending',()=>{
 expect(providerActivation('screenscraper',true,false)).toBe(false);
 expect(providerActivation('screenscraper',true,true)).toBe(true);
 expect(providerActivation('screenscraper',false,true)).toBe(false);
 expect(providerActivation('igdb',true,false)).toBe(false);
});
