import {BEAT_SECONDS} from './music/composition';
export function ambientMotion(seconds:number,level:number,notePulse=0){
 const phase=((seconds%24)+24)%24;
 const breath=.5+.5*Math.sin(2*Math.PI*seconds/(BEAT_SECONDS*8));
 const pulse=Math.max(0,Math.min(1,notePulse));
 return{phase,breath,pulse,intensity:Math.min(1,.16+breath*.2+Math.max(0,level)*.28+pulse*.25)};
}
