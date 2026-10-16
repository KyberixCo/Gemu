import {libraryHost,type Entry,type Snapshot} from '../../library';
export type RetroSystem={id:string;name:string;core:string;coreName:string;extensions:string[];firmware:{name:string;md5:string;region:string}[]|null};
export type RetroCollection={systemId:string;directory:string;recursive:boolean};
export type RetroConfig={executable:string;coreDirectory?:string;systemDirectory:string;collections:RetroCollection[];fullscreen:boolean;rewind:boolean;language:string;joypadDriver:string};
export type RetroStatus={supported:boolean;installations:{executable:string;managed:boolean}[];config:RetroConfig;systems:RetroSystem[];diagnostics:string[];defaultDirectories:Record<string,string>};
export type RetroPlan={token:string;version:string;destination:string;recommendedBytes?:number;freeBytes?:number;downloads:{name:string;url:string;size:number}[]};
export type RetroJob={id:string;state:string;current:string;bytes:number;total:number;completed:number;count:number;executable:string;error:string};
export type RetroPreview={token:string;games:Entry[];checks:{systemId:string;corePath:string;coreReady:boolean;firmware:{name:string;region:string;state:string}[]}[];diagnostics:string[];canPublish:boolean};
export type RetroHost={
 GetRetroArchStatus:(profileID:string,manual:string)=>Promise<RetroStatus>;
 PlanRetroArchInstall:(profileID:string,request:{executable:string;coreDirectory?:string;installApp:boolean;systems:string[]})=>Promise<RetroPlan>;
 DiscardRetroArchInstallPlan:(profileID:string,token:string)=>Promise<void>;
 StartRetroArchInstall:(profileID:string,token:string)=>Promise<RetroJob>;
 GetRetroArchInstall:(profileID:string,id:string)=>Promise<RetroJob>;
 CancelRetroArchInstall:(profileID:string,id:string)=>Promise<void>;
 PreviewRetroArch:(profileID:string,config:RetroConfig)=>Promise<RetroPreview>;
 DiscardRetroArchPreview:(profileID:string,token:string)=>Promise<void>;
 CommitRetroArch:(profileID:string,token:string)=>Promise<Snapshot>;
 PickRetroArchPath:(profileID:string,kind:string)=>Promise<string>;
 CreateRetroArchROMFolders:(profileID:string,systems:string[])=>Promise<Record<string,string>>;
 OpenRetroArchMenu:(profileID:string,config:RetroConfig)=>Promise<unknown>;
};
export function retroHost():RetroHost{
 const host=libraryHost() as ReturnType<typeof libraryHost>&RetroHost;
 if(['GetRetroArchStatus','PlanRetroArchInstall','StartRetroArchInstall','PreviewRetroArch','CommitRetroArch','OpenRetroArchMenu'].some(key=>typeof host[key as keyof RetroHost]!=='function'))throw new Error('Reinicia Gemu para instalar y configurar RetroArch.');
 return host;
}
export function isInstalling(job:RetroJob|undefined){return Boolean(job&&!['ready','error','cancelled'].includes(job.state));}
export function chooseCollections(previous:readonly RetroCollection[],systems:readonly string[]):RetroCollection[]{return systems.map(systemId=>previous.find(item=>item.systemId===systemId)??{systemId,directory:'',recursive:true});}
export function installSize(plan:RetroPlan){return plan.downloads.every(item=>item.size>=0)?plan.downloads.reduce((total,item)=>total+item.size,0):null;}
