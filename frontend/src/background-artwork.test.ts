import {afterEach,expect,it,vi} from 'vitest';
import {isBackgroundSize,prepareBackgroundArtwork} from './background-artwork';

afterEach(()=>vi.unstubAllGlobals());

it('rejects small Steam headers, undersized landscapes and portrait covers',()=>{
 for(const [width,height] of [[460,215],[1232,706],[1920,600],[600,900],[1280,1920],[0,0]]){
  expect(isBackgroundSize(width,height)).toBe(false);
 }
 expect(isBackgroundSize(1280,720)).toBe(true);
 expect(isBackgroundSize(1920,1080)).toBe(true);
});

it('checks decoded image dimensions before admitting artwork to the background rotation',async()=>{
 let width=460,height=215;
 vi.stubGlobal('Image',class{
  src='';naturalWidth=width;naturalHeight=height;
  decode(){return Promise.resolve();}
 });
 const artwork={id:'background',type:'image' as const,src:'/library-assets/cached.img'};
 expect(await prepareBackgroundArtwork(artwork)).toBe(false);
 width=1920;height=1080;
 expect(await prepareBackgroundArtwork(artwork)).toBe(true);
});

it('rejects broken images and retains procedural scenery',async()=>{
 vi.stubGlobal('Image',class{src='';decode(){return Promise.reject(new Error('Invalid image'));}});
 expect(await prepareBackgroundArtwork({id:'broken',type:'image',src:'/broken.img'})).toBe(false);
 expect(await prepareBackgroundArtwork({id:'fallback',type:'procedural',variant:0})).toBe(true);
});
