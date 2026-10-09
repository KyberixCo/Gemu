type SceneAudio={playing:boolean;level:number;pulse:number;crystal:number;wind:number;warmth:number;pan:number};
const clamp=(value:number,min=0,max=1)=>Math.max(min,Math.min(max,Number.isFinite(value)?value:0));
export function sceneLighting(audio:SceneAudio){
 if(!audio.playing)return{brightness:1,saturation:1,hue:0,glow:0,air:0,position:74};
 // Quiet musical cues need a perceptual response to remain visible behind the scene shade.
 const level=Math.sqrt(clamp(audio.level)),pulse=Math.sqrt(clamp(audio.pulse)),crystal=Math.sqrt(clamp(audio.crystal));
 const wind=Math.sqrt(clamp(audio.wind)),tone=clamp(audio.warmth,-1,1);
 const warmth=Math.sign(tone)*Math.sqrt(Math.abs(tone));
 return{brightness:1+level*.35+pulse*.22+wind*.06,saturation:1+level*.4+crystal*.28+Math.abs(warmth)*.18,
  hue:warmth*28,glow:crystal*.62+pulse*.12,air:wind*.3,position:74+clamp(audio.pan,-1,1)*18};
}
