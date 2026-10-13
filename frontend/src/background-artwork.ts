import type {GameArtwork} from './catalog';

// Full-screen scenery must not enlarge small headers or portrait covers.
export function isBackgroundSize(width:number,height:number){
 return width>=1280&&height>=720&&width>height;
}

export async function prepareBackgroundArtwork(artwork:GameArtwork){
 if(artwork.type==='procedural')return true;
 const image=new Image();image.src=artwork.src;
 let timer:ReturnType<typeof setTimeout>|undefined;
 try{
  return await Promise.race([
   image.decode().then(()=>isBackgroundSize(image.naturalWidth,image.naturalHeight),()=>false),
   new Promise<boolean>(resolve=>{timer=setTimeout(()=>resolve(false),4000);}),
  ]);
 }finally{clearTimeout(timer);}
}
