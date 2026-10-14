import {afterEach,expect,it,vi} from 'vitest';
import {toggleFullscreen} from './fullscreen';

afterEach(()=>vi.unstubAllGlobals());
it.each([false,true])('toggles the native host when fullscreen is %s',async fullscreen=>{
 const desktop={WindowIsFullscreen:vi.fn().mockResolvedValue(fullscreen),WindowFullscreen:vi.fn(),WindowUnfullscreen:vi.fn()};
 vi.stubGlobal('window',{runtime:desktop});
 // No document: the native path must never enter browser fullscreen.
 await toggleFullscreen();
 expect(desktop.WindowFullscreen).toHaveBeenCalledTimes(fullscreen?0:1);
 expect(desktop.WindowUnfullscreen).toHaveBeenCalledTimes(fullscreen?1:0);
});
it.each([false,true])('keeps browser preview fullscreen working when fullscreen is %s',async fullscreen=>{
 const enter=vi.fn(),exit=vi.fn();
 vi.stubGlobal('window',{});
 vi.stubGlobal('document',{fullscreenElement:fullscreen?{}:null,documentElement:{requestFullscreen:enter},exitFullscreen:exit});
 await toggleFullscreen();
 expect(enter).toHaveBeenCalledTimes(fullscreen?0:1);
 expect(exit).toHaveBeenCalledTimes(fullscreen?1:0);
});
it('propagates unavailable fullscreen so the UI can report it',async()=>{
 vi.stubGlobal('window',{});
 vi.stubGlobal('document',{fullscreenElement:null,documentElement:{requestFullscreen:vi.fn().mockRejectedValue(new Error('unavailable'))}});
 await expect(toggleFullscreen()).rejects.toThrow('unavailable');
});
