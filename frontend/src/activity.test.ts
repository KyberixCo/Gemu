import {afterEach,expect,it,vi} from 'vitest';
import {isForeground,isGameplay,setAssisting,setGameplay,subscribeActivity} from './activity';
afterEach(()=>{setGameplay(false);setAssisting(false);vi.unstubAllGlobals();});
it('enters low power during gameplay and restores foreground activity',()=>{
 const document={hidden:false,hasFocus:()=>true,documentElement:{dataset:{} as Record<string,string>}};
 vi.stubGlobal('document',document);
 const changed=vi.fn(),unsubscribe=subscribeActivity(changed);
 expect(isForeground()).toBe(true);setGameplay(true);
 expect(isGameplay()).toBe(true);expect(isForeground()).toBe(false);expect(document.documentElement.dataset.lowPower).toBe('true');
 setGameplay(true);expect(changed).toHaveBeenCalledTimes(1);
 setGameplay(false);expect(isForeground()).toBe(true);expect(changed).toHaveBeenCalledTimes(2);
 document.hidden=true;expect(isForeground()).toBe(false);
 unsubscribe();
});
it('pauses local controller navigation while desktop assistance owns input',()=>{
 vi.stubGlobal('document',{hidden:false,hasFocus:()=>true});
 const changed=vi.fn(),stop=subscribeActivity(changed);
 setAssisting(true);expect(isForeground()).toBe(false);
 setAssisting(true);expect(changed).toHaveBeenCalledTimes(1);
 setAssisting(false);expect(isForeground()).toBe(true);expect(changed).toHaveBeenCalledTimes(2);
 stop();
});
