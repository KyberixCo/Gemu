import {expect,it} from 'vitest';
import {shouldOfferSetup} from './policy';
it('offers first-time setup only for incomplete profiles without a published library',()=>{
 expect(shouldOfferSetup(false,{sources:[],games:[]})).toBe(true);
 expect(shouldOfferSetup(true,{sources:[],games:[]})).toBe(false);
 expect(shouldOfferSetup(false,{sources:[{id:'steam',root:'',profileId:''}],games:[]})).toBe(false);
});
