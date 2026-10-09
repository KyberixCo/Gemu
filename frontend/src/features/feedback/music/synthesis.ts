import {random} from './profile';
import type {MusicNote} from './composition';
export const isAirTexture=(instrument:MusicNote['instrument'])=>['wind','gust','surf','rustle'].includes(instrument);
export function instrumentBuffers(context:AudioContext) {
 const rate=context.sampleRate,base=261.625565;
 const samples={} as Record<MusicNote['instrument'],AudioBuffer>;
 for(const instrument of ['pad','bell','pluck','wind','crystal','bass','flute','mallet','chip','gust','surf','rustle'] as const){
  const length=instrument==='pad'?Math.round(rate*128/base):Math.round(rate*(isAirTexture(instrument)?8:instrument==='crystal'?6:4));
  const buffer=context.createBuffer(2,length,rate);
  for(let channel=0;channel<2;channel++){
   const data=buffer.getChannelData(channel);
   const rng=random(7391+channel*113);let slow=0,fast=0;
   for(let i=0;i<length;i++){
    const t=i/rate,phase=instrument==='pad'?2*Math.PI*128*i/length:2*Math.PI*base*t;
    if(instrument==='pad')data[i]=.28*(Math.sin(phase+channel*.16)+.055*Math.sin(phase*2)+.022*Math.sin(phase*3));
    else if(isAirTexture(instrument)){
     const noise=rng()*2-1;
     // Band-limited noise: remove low rumble and soften the high-frequency hiss.
     const low=instrument==='surf'?65:instrument==='rustle'?650:100;
     const high=instrument==='gust'?700+1300*(.5+.5*Math.sin(t*.9+channel*.4)):instrument==='rustle'?2800:1100;
     slow+=(1-Math.exp(-2*Math.PI*low/rate))*(noise-slow);
     fast+=(1-Math.exp(-2*Math.PI*high/rate))*(noise-fast);
     const envelope=Math.sin(Math.PI*i/(length-1))**2;
     const motion=instrument==='rustle'?.45+.55*Math.sin(t*13+Math.sin(t*3)+channel*.9)**2
      :instrument==='surf'?.35+.65*Math.sin(t*1.1+channel*.3)**2
      :.8+.2*Math.sin(t*1.7+channel*.8);
     data[i]=(fast-slow)*.65*envelope*motion;
    }else if(instrument==='crystal'){
     const attack=Math.min(1,t/.008),end=Math.min(1,(length-1-i)/(.08*rate));
     // Glass-like resonances, with a tuned fundamental and quickly fading overtones.
     data[i]=.24*attack*Math.max(0,end)*(Math.sin(phase)*Math.exp(-t*1.5)+.28*Math.sin(phase*2.76+channel*.08)*Math.exp(-t*3.8)+.12*Math.sin(phase*5.4)*Math.exp(-t*6));
    }else if(instrument==='bass'||instrument==='flute'||instrument==='mallet'||instrument==='chip'){
     const attack=Math.min(1,t/(instrument==='flute'?.12:.015));
     const end=Math.max(0,Math.min(1,(length-1-i)/(.08*rate)));
     const decay=Math.exp(-t*(instrument==='flute'?.65:instrument==='bass'?1.1:2.5));
     // Band-limited harmonic sums keep the retro pulse soft and avoid square-wave aliasing.
     const tone=instrument==='chip'?Math.sin(phase)+.28*Math.sin(phase*3)+.12*Math.sin(phase*5)
      :instrument==='flute'?Math.sin(phase)+.12*Math.sin(phase*2+Math.sin(t*5)*.2)
      :instrument==='mallet'?Math.sin(phase)+.35*Math.sin(phase*4)*Math.exp(-t*6)
      :Math.sin(phase)+.16*Math.sin(phase*2);
     data[i]=.28*attack*end*decay*tone;
    }else{
     const decay=Math.exp(-t*(instrument==='bell'?1.8:3.2)),attack=Math.min(1,t/.012),end=Math.min(1,(length-1-i)/(.06*rate));
     data[i]=.32*attack*decay*Math.max(0,end)*(Math.sin(phase)+.17*Math.sin(phase*2.006)*Math.exp(-t*3)+.05*Math.sin(phase*3.97)*Math.exp(-t*5));
    }
   }
  }
  samples[instrument]=buffer;
 }
 return samples;
}
export function roomImpulse(context:AudioContext) {
 const length=Math.round(context.sampleRate*2.6),buffer=context.createBuffer(2,length,context.sampleRate);
 for(let channel=0;channel<2;channel++){
  const values=buffer.getChannelData(channel),rng=random(1947+channel*87);let filtered=0;
  for(let i=0;i<length;i++){
   const t=i/context.sampleRate;filtered+=.22*((rng()*2-1)-filtered);
   values[i]=filtered*Math.exp(-t*3.3)*Math.min(1,t/.025)*Math.min(1,(length-1-i)/(context.sampleRate*.08));
  }
 }
 return buffer;
}
