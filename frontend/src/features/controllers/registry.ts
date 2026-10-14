import type {ControllerDevice} from './providers';
export const controllerRegistryKey='gemu.controllers.v1';
export const uuidPattern=/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
export type RegisteredController={uuid:string;name:string;model:string;playerId:string;provider:string;stableId?:string};
export function readControllerRegistry(value:unknown):RegisteredController[]{
 if(!Array.isArray(value))return [];
 const seen=new Set<string>();
 return value.filter((item):item is RegisteredController=>{
  if(!item||typeof item!=='object'||!uuidPattern.test(item.uuid)||seen.has(item.uuid)||typeof item.name!=='string'||!item.name.trim()||item.name.length>64||['model','playerId','provider'].some(key=>typeof item[key]!=='string')||(item.stableId!==undefined&&typeof item.stableId!=='string'))return false;
  seen.add(item.uuid);return true;
 });
}
export function newRegisteredController(device:ControllerDevice,uuid=crypto.randomUUID()):RegisteredController{
 return {uuid,name:device.name.slice(0,64)||'Gamepad',model:device.name,provider:device.provider,playerId:'',...(device.stableId?{stableId:device.stableId}:{})};
}
// Connections have session identities. Never infer a physical device from its model or slot.
export class ControllerAssociations{
 private bindings=new Map<string,string>();
 private key(device:ControllerDevice){return `${device.provider}:${device.id}`;}
 observe(devices:readonly ControllerDevice[]){const live=new Set(devices.map(device=>this.key(device)));for(const key of this.bindings.keys())if(!live.has(key))this.bindings.delete(key);}
 resolve(device:ControllerDevice,records:readonly RegisteredController[]){
  const associated=this.bindings.get(this.key(device));
  if(associated)return records.find(record=>record.uuid===associated);
  if(device.stableId){const matches=records.filter(record=>record.provider===device.provider&&record.stableId===device.stableId);if(matches.length===1)return matches[0];}
  return undefined;
 }
 attach(device:ControllerDevice,record:RegisteredController){
  for(const [key,uuid] of this.bindings)if(uuid===record.uuid&&key!==this.key(device))throw new Error('Este mando guardado ya está asociado a otro dispositivo conectado.');
  this.bindings.set(this.key(device),record.uuid);
 }
}
export const controllerAssociations=new ControllerAssociations();
