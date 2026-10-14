import {detectControllerFamily} from '../management/controller-devices';
import type {Family} from '../management/controller-profiles';

export type Capability={supported:boolean;reason?:string};
export type ControllerDevice={provider:string;id:string;stableId?:string;name:string;family:Family|null;playerNumber?:number;capabilities:Record<string,Capability>;battery:{state:string;level?:string;percent?:number}};
export type ControllerFeatureRequest={provider:string;deviceId:string;feature:string;value:unknown};
export interface ControllerProvider{
 readonly id:string;
 devices():Promise<ControllerDevice[]>;
 setFeature(request:ControllerFeatureRequest):Promise<void>;
}
export function browserController(index:number,id:string):ControllerDevice{
 return {provider:'browser',id:`${index}:${id}`,name:id,family:detectControllerFamily(id),capabilities:{identity:{supported:false},battery:{supported:false},'player-led':{supported:false}},battery:{state:'unsupported'}};
}
export function browserControllerProvider():ControllerProvider{
 return {id:'browser',async devices(){
  const pads=navigator.getGamepads?.()??[];
  return pads.filter((pad):pad is Gamepad=>!!pad?.connected).map(pad=>browserController(pad.index,pad.id));
 },async setFeature(){throw new Error('Función no disponible.');}};
}
type NativeHost={GetControllers:()=>Promise<ControllerDevice[]>;SetControllerFeature:(request:ControllerFeatureRequest)=>Promise<void>};
export function nativeControllerProvider():ControllerProvider{
 const host=()=> (window as Window&{go?:{main?:{App?:NativeHost}}}).go?.main?.App;
 return {id:'native',async devices(){return host()?.GetControllers?.()??[];},async setFeature(request){
  const app=host();if(!app?.SetControllerFeature)throw new Error('Función no disponible.');await app.SetControllerFeature(request);
 }};
}
export class ControllerProviders{
 private providers=new Map<string,ControllerProvider>();
 register(provider:ControllerProvider){if(this.providers.has(provider.id))throw new Error('Duplicate controller provider');this.providers.set(provider.id,provider);}
 async devices(){
  const results=await Promise.allSettled([...this.providers.values()].map(provider=>provider.devices()));
  return {devices:results.flatMap(result=>result.status==='fulfilled'?result.value:[]),errors:results.flatMap(result=>result.status==='rejected'?[String(result.reason)]:[])};
 }
 async setFeature(device:ControllerDevice,feature:string,value:unknown){
  if(!device.capabilities[feature]?.supported)throw new Error('Función no disponible.');
  const provider=this.providers.get(device.provider)??this.providers.get('native');
  if(!provider)throw new Error('Función no disponible.');
  await provider.setFeature({provider:device.provider,deviceId:device.id,feature,value});
 }
}
