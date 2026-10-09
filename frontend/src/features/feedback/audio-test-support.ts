import {vi} from 'vitest';
export class FakeParam{
 value=1;cancelScheduledValues=vi.fn();setTargetAtTime=vi.fn();setValueAtTime=vi.fn();linearRampToValueAtTime=vi.fn();
}
class FakeNode{connect=vi.fn();disconnect=vi.fn();}
export class FakeGain extends FakeNode{gain=new FakeParam();}
export class FakeSource extends FakeNode{
 buffer:unknown;onended?:()=>void;loop=false;playbackRate=new FakeParam();start=vi.fn();stop=vi.fn();
}
class FakeAnalyser extends FakeNode{fftSize=512;getFloatTimeDomainData(data:Float32Array){data.fill(.002);}}
class FakeFilter extends FakeNode{type='lowpass';frequency=new FakeParam();}
class FakePan extends FakeNode{pan=new FakeParam();}
class FakeConvolver extends FakeNode{buffer:unknown;normalize=true;}
class FakeDelay extends FakeNode{delayTime=new FakeParam();}
export class FakeContext{
 static instances:FakeContext[]=[];
 state='running';currentTime=5;sampleRate=8000;destination={};
 sources:FakeSource[]=[];gains:FakeGain[]=[];analysers:FakeAnalyser[]=[];filters:FakeFilter[]=[];
 constructor(){FakeContext.instances.push(this);}
 resume=vi.fn(async()=>{this.state='running';});decodeAudioData=vi.fn(async()=>({duration:24}));
 createAnalyser(){const node=new FakeAnalyser();this.analysers.push(node);return node;}
 createGain(){const node=new FakeGain();this.gains.push(node);return node;}
 createBufferSource(){const node=new FakeSource();this.sources.push(node);return node;}
 createBiquadFilter(){const node=new FakeFilter();this.filters.push(node);return node;}
 createStereoPanner(){return new FakePan();}createConvolver(){return new FakeConvolver();}createDelay(){return new FakeDelay();}
 createBuffer(channels:number,length:number,sampleRate:number){const data=Array.from({length:channels},()=>new Float32Array(length));return{duration:length/sampleRate,getChannelData:(channel:number)=>data[channel]};}
}
