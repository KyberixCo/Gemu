import {useEffect,useMemo,useState} from 'react';
import {libraryHost,type Snapshot} from './library';
import {LibrarySearch,emptySearch,type SearchHost,type SearchView} from './search';

export function useLibrarySearch({profileID,snapshot,demo,query,sourceID,favoritesOnly,favorites}:{profileID:string;snapshot:Snapshot;demo:boolean;query:string;sourceID:string;favoritesOnly:boolean;favorites:readonly string[]}){
 const key=JSON.stringify([profileID,query,sourceID,favoritesOnly,favorites]);
 const [result,setResult]=useState<{key:string;snapshot:Snapshot;view:SearchView}>();
 const controller=useMemo(()=>new LibrarySearch((id,request)=>{
  const host=libraryHost() as ReturnType<typeof libraryHost>&SearchHost;
  if(typeof host.SearchLibrary!=='function')throw new Error('Reinicia Gemu para usar la búsqueda de biblioteca.');
  return host.SearchLibrary(id,request);
 },view=>setResult({key,snapshot,view})),[key,snapshot]);
 const enabled=!demo&&Boolean(query.trim());
 useEffect(()=>{if(enabled)controller.search(profileID,{query,sourceId:sourceID,favoritesOnly,limit:48,offset:0});return()=>controller.cancel();},[controller,enabled]);
 const view=result?.key===key&&result.snapshot===snapshot?result.view:{...emptySearch,status:'loading' as const};
 return {...view,enabled,more:()=>controller.more(),retry:()=>controller.retry()};
}
