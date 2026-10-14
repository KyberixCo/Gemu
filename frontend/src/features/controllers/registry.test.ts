import {describe,it,expect} from 'vitest';
import {ControllerAssociations,newRegisteredController,readControllerRegistry} from './registry';
import {browserController,ControllerProviders,type ControllerProvider} from './providers';
import {defaultProfile,validProfiles} from '../management/controller-profiles';
const uuid='11111111-2222-4333-8444-555555555555';
describe('saved controller identities',()=>{
 it('keeps identical models separate and requires reassociation after disconnect',()=>{
  const first=browserController(0,'Xbox'),second=browserController(1,'Xbox'),record=newRegisteredController(first,uuid),associations=new ControllerAssociations();
  associations.attach(first,record);expect(associations.resolve(second,[record])).toBeUndefined();
  expect(()=>associations.attach(second,record)).toThrow();associations.observe([second]);
  expect(associations.resolve(first,[record])).toBeUndefined();associations.attach(second,record);
  expect(associations.resolve(second,[record])?.uuid).toBe(uuid);
 });
 it('recovers a unique persistent identity from a capable provider',()=>{
  const device={...browserController(0,'model'),stableId:'serial-a'},record=newRegisteredController(device,uuid),associations=new ControllerAssociations();
  expect(associations.resolve({...device,id:'other-session'},[record])?.uuid).toBe(uuid);
  expect(associations.resolve({...device,provider:'another'},[record])).toBeUndefined();
  expect(associations.resolve(device,[record,{...record,uuid:'aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee'}])).toBeUndefined();
 });
 it('validates records and preserves UUID mappings alongside legacy family profiles',()=>{
  const record=newRegisteredController(browserController(0,'Xbox'),uuid);
  expect(readControllerRegistry([record,record,{...record,uuid:'bad'}])).toEqual([record]);
  expect(validProfiles({[`Global:xbox:${uuid}`]:defaultProfile(),'Global:xbox':defaultProfile()})).toBe(true);
  expect(validProfiles({'Global:xbox:slot-0':defaultProfile()})).toBe(false);
 });
});
describe('extensible controller providers',()=>{
 it('isolates failures and dispatches only supported features',async()=>{
  const service=new ControllerProviders(),device=browserController(0,'Xbox');let calls=0;
  const provider:ControllerProvider={id:'browser',devices:async()=>[device],setFeature:async()=>{calls++;}};
  service.register(provider);service.register({id:'broken',devices:async()=>{throw Error('offline');},setFeature:async()=>{}});
  expect(()=>service.register(provider)).toThrow();const result=await service.devices();expect(result.devices).toEqual([device]);expect(result.errors).toHaveLength(1);
  await expect(service.setFeature(device,'player-led',2)).rejects.toThrow();expect(calls).toBe(0);
  await service.setFeature({...device,capabilities:{'player-led':{supported:true}}},'player-led',2);expect(calls).toBe(1);
 });
});
