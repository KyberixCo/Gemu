import type {ContentRequest} from '../../library';
import type {ContentSettings} from './ContentProviders';

export function contentRequest(id:ContentRequest['id'],settings:ContentSettings,credentials:Record<string,string>,language:'es'|'en'):ContentRequest{
 const names=id==='steamgriddb'?['apiKey']:['username','password'];
 return {id,priority:Number(settings[id].priority),language,credentials:Object.fromEntries(names.map((name,index)=>[name,credentials[`${id}:${index}`]??'']))};
}
export function contentRequests(settings:ContentSettings,credentials:Record<string,string>,language:'es'|'en'):ContentRequest[]{
 return (['steamgriddb','screenscraper'] as const).filter(id=>settings[id].enabled).map(id=>contentRequest(id,settings,credentials,language)).sort((a,b)=>a.priority-b.priority);
}
