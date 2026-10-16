import {preferencesStorage} from '../../persistence';
import {profileKey} from '../../profiles';
import {defaultContentSettings,readContentSettings,type ContentSettings} from './content-providers';
import {validProfiles,type ControllerProfiles} from './controller-profiles';
export const providers = {
 minimax: { name:'MiniMax · M3', endpoint:'https://api.minimax.io/v1', model:'MiniMax-M3', cost:'Según consumo' },
 openrouter: { name:'OpenRouter · gratuitos', endpoint:'https://openrouter.ai/api/v1', model:'openrouter/free', cost:'Gratis · sujeto a límites' },
 ollama: { name:'Ollama · local', endpoint:'http://localhost:11434/v1', model:'', cost:'Sin coste de API · requiere equipo local' },
 compatible: { name:'API compatible', endpoint:'', model:'', cost:'Configura tu proveedor de bajo coste' },
};
export type Provider = keyof typeof providers;
export type Config = { contentProviders:ContentSettings; controllers:ControllerProfiles; sources:Record<string,{enabled:boolean;path:string}>; emulators:Record<string,{binary:string;roms:string;screen:string;core:string;shader:string;rewind:boolean}>; metadata:Record<string,{title:string;genre:string;description:string;cover:string;background:string;logo:string}>; llm:{provider:Provider;endpoint:string;model:string;budget:string} };
export const defaults:Config = {
 contentProviders:defaultContentSettings(), controllers:{},
 sources:Object.fromEntries(['Steam','Epic Games','GOG','Manual','RetroArch'].map(name=>[name,{enabled:false,path:''}])),
 emulators:Object.fromEntries(['RetroArch','Dolphin','PCSX2','PPSSPP'].map(name=>[name,{binary:'',roms:'',screen:'Pantalla completa',core:'',shader:'Ninguno',rewind:false}])),
 metadata:{}, llm:{provider:'minimax',endpoint:providers.minimax.endpoint,model:providers.minimax.model,budget:'2'},
};
export function loadManagementConfig():Config {try {const value=JSON.parse(preferencesStorage.getItem(profileKey('gemu.management.v1'))||'null');if(value && Object.keys(defaults.sources).filter(key=>key!=='RetroArch').every(k=>typeof value.sources?.[k]?.path==='string'&&typeof value.sources[k].enabled==='boolean') && Object.keys(defaults.emulators).every(k=>typeof value.emulators?.[k]?.binary==='string'&&typeof value.emulators[k].roms==='string'&&typeof value.emulators[k].screen==='string'&&typeof value.emulators[k].core==='string'&&typeof value.emulators[k].shader==='string'&&typeof value.emulators[k].rewind==='boolean') && value.metadata && typeof value.metadata==='object' && !Array.isArray(value.metadata) && Object.values(value.metadata).every(entry=>entry && typeof entry==='object' && ['title','genre','description','cover','background','logo'].every(field=>typeof (entry as Record<string,unknown>)[field]==='string')) && value.llm?.provider in providers && ['endpoint','model','budget'].every(k=>typeof value.llm[k]==='string')) return {...value,sources:readSourceSettings(value.sources),contentProviders:readContentSettings(value.contentProviders),controllers:validProfiles(value.controllers)?value.controllers:{}};}catch{/* Use demo defaults if storage is unavailable. */}return structuredClone(defaults);}

export function saveManagementConfig(config:Config){return preferencesStorage.setItem(profileKey('gemu.management.v1'),JSON.stringify(config));}

function readSourceSettings(value:Record<string,unknown>):Config['sources']{
 const result=structuredClone(defaults.sources);
 for(const [key,item] of Object.entries(value))if(item&&typeof item==='object'&&typeof (item as {enabled?:unknown}).enabled==='boolean'&&typeof (item as {path?:unknown}).path==='string')result[key]={enabled:(item as {enabled:boolean}).enabled,path:(item as {path:string}).path};
 return result;
}