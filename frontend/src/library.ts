import type {Game} from './catalog';
import {isBackgroundSize} from './background-artwork';
export type Profile={id:string;name:string;origin:string};
export type SourceInfo={id:string;name:string};
export type Achievement={id:string;name:string;description:string;unlocked:boolean;unlockedAt:number};
export type LibraryArtwork={id:string;kind:'background'|'cover'|'logo'|'icon';url:string;provider:string;author?:string;page?:string;width:number;height:number};
export type ContentRequest={id:'steamgriddb'|'screenscraper';priority:number;credentials:Record<string,string>};
export type CredentialProvider='steamgriddb'|'screenscraper'|'igdb'|'steam'|'minimax'|'openrouter'|'compatible';
export type CredentialStorage={name:string;available:boolean};
export type CredentialStatus={provider:CredentialProvider;savedFields:string[];storage:CredentialStorage};
export type PlayerState={profile:Profile;status:string;message:string;achievements:Achievement[]};
export type InstalledGame={ID:string;Title:string;Installations:{Provider:string;ExternalID:string;Path:string;Installed:boolean}[]};
export type Entry={game:InstalledGame;sourceId:string;metadata:{description:string;genres:string[];cover:string;background:string;origin:string;status:string;logo?:string;icon?:string;backgrounds?:string[];manualArt?:string[];artworks?:LibraryArtwork[]};player:PlayerState};
export type SourceConfig={id:string;name?:string;root:string;profileId:string};
export type Snapshot={version:number;sources:SourceConfig[];games:Entry[];artworkUndo?:unknown};
export type Preview={root:string;games:InstalledGame[];profiles:Profile[];diagnostics:string[]};
export type Draft={token:string;source:SourceConfig;games:Entry[];diagnostics:string[]};
export type ArtworkProposal={gameId:string;title:string;candidates:LibraryArtwork[];diagnostics:string[]};
export type ArtworkSearch={id:string;state:string;total:number;completed:number;current:string;proposals:ArtworkProposal[]};
export type ArtworkChoice={gameId:string;ids:string[];rotateIds?:string[]};
type Bridge={StartArtworkSearch:(request:{gameIds:string[];providers:ContentRequest[];missingOnly:boolean})=>Promise<ArtworkSearch>;GetArtworkSearch:(id:string)=>Promise<ArtworkSearch>;CancelArtworkSearch:(id:string)=>Promise<void>;DiscardArtworkSearch:(id:string)=>Promise<void>;CommitArtwork:(request:{id:string;choices:ArtworkChoice[]})=>Promise<Snapshot>;UndoArtwork:()=>Promise<Snapshot>;GetCredentialStatus:(provider:CredentialProvider)=>Promise<CredentialStatus>;SaveProviderCredentials:(request:{provider:CredentialProvider;fields:Record<string,string>})=>Promise<void>;DeleteProviderCredentials:(provider:CredentialProvider)=>Promise<void>;GetSources:()=>Promise<SourceInfo[]>;CheckContentProvider:(request:ContentRequest)=>Promise<void>;PreviewSource:(sourceId:string,root:string)=>Promise<Preview>;PrepareLibraryImport:(request:{sourceId:string;root:string;profileId:string;apiKey:string;contentProviders?:ContentRequest[]})=>Promise<Draft>;CommitLibraryImport:(token:string)=>Promise<Snapshot>;DiscardLibraryImport:(token:string)=>Promise<void>;GetLibrary:()=>Promise<Snapshot>};
export function libraryHost():Bridge {
 const app=(window as Window & {go?:{main?:{App?:Bridge}}}).go?.main?.App;
 if(!app)throw new Error('Abre la aplicación de escritorio para configurar tu biblioteca.');
 return app;
}
export function contentHost(host=libraryHost()):Bridge {
 if(typeof host.CheckContentProvider!=='function')throw new Error('Esta ventana de Gemu usa una versión anterior del host. Cierra Gemu y reinicia dev.ps1, o abre el ejecutable actualizado. Recargar la interfaz no actualiza el host. Las credenciales de sesión deberán introducirse de nuevo.');
 return host;
}
export function credentialHost(host=libraryHost()):Bridge {
 if(['GetCredentialStatus','SaveProviderCredentials','DeleteProviderCredentials'].some(name=>typeof host[name as keyof Bridge]!=='function'))throw new Error('Reinicia Gemu con dev.ps1 o abre el ejecutable actualizado para usar el gestor seguro de credenciales. Recargar la interfaz no actualiza el host.');
 return host;
}
export async function loadLibrarySession(host:Pick<Bridge,'GetSources'|'GetLibrary'>,callbacks:{sources:(sources:SourceInfo[])=>void;library:(snapshot:Snapshot)=>void;error:(part:'sources'|'library',error:unknown)=>void}) {
 await Promise.allSettled([
  Promise.resolve().then(()=>{
   if(typeof host.GetSources!=='function')throw new Error('El host no tiene los métodos actuales. Cierra Gemu y vuelve a ejecutar dev.ps1.');
   return host.GetSources();
  }).then(callbacks.sources).catch(error=>callbacks.error('sources',error)),
  Promise.resolve().then(()=>host.GetLibrary()).then(callbacks.library).catch(error=>callbacks.error('library',error)),
 ]);
}
export function catalogFromLibrary(snapshot:Snapshot):Game[] {
 return snapshot.games.map(entry=>{
  const art=[];
  if(entry.metadata.background)art.push({id:'background',type:'image' as const,src:entry.metadata.background});
  for(const artwork of entry.metadata.artworks??[]){
   if(((artwork.kind==='background'&&(!entry.metadata.backgrounds||entry.metadata.backgrounds.includes(artwork.url)))||(artwork.kind==='cover'&&entry.metadata.backgrounds?.includes(artwork.url)&&isBackgroundSize(artwork.width,artwork.height)))&&!art.some(existing=>existing.src===artwork.url))art.push({id:artwork.id,type:'image' as const,src:artwork.url});
  }
  // The selected cover can also be scenery when its actual dimensions qualify.
  const cover=entry.metadata.artworks?.find(a=>a.url===entry.metadata.cover&&isBackgroundSize(a.width,a.height));
  if(cover&&!art.some(a=>a.src===cover.url))art.push({id:'scenery-cover',type:'image' as const,src:cover.url});
  return {id:entry.game.ID,title:entry.game.Title,provider:snapshot.sources.find(source=>source.id===entry.sourceId)?.name||entry.sourceId,genre:entry.metadata.genres?.join(' · ')||'Sin clasificar',subtitle:entry.player.profile.name?`Perfil · ${entry.player.profile.name}`:'Biblioteca instalada',description:entry.metadata.description||'Metadatos pendientes de obtener.',color:'#627f91',symbol:'◉',platform:'PC',installed:entry.game.Installations.some(i=>i.Installed),player:entry.player,cover:entry.metadata.cover,logo:entry.metadata.logo,icon:entry.metadata.icon,artworks:art.length?art:undefined};
 });
}

export function artworkHost(host=libraryHost()){if(['StartArtworkSearch','GetArtworkSearch','CancelArtworkSearch','DiscardArtworkSearch','CommitArtwork','UndoArtwork'].some(name=>typeof host[name as keyof Bridge]!=="function"))throw new Error("Reinicia Gemu con dev.ps1 para usar la búsqueda de arte. Recargar la interfaz no actualiza el host.");return host;}
