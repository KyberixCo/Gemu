export type SearchRequest={query:string;sourceId:string;favoritesOnly:boolean;limit:number;offset:number};
export type SearchHit={gameId:string;title:string;sourceId:string;kind:string;score:number};
export type SearchPage={hits:SearchHit[];total:number;hasMore:boolean};
export type SearchHost={SearchLibrary:(profileID:string,request:SearchRequest)=>Promise<SearchPage>};
export type SearchView=SearchPage&{status:'idle'|'loading'|'ready'|'error';error:string};
export const emptySearch:SearchView={hits:[],total:0,hasMore:false,status:'idle',error:''};

// Wails requests cannot be aborted, so every completion is checked against the
// current generation, including errors and requests from a previous player.
export class LibrarySearch {
 private generation=0;
 private timer:ReturnType<typeof setTimeout>|undefined;
 private current:{profileID:string;request:SearchRequest}|undefined;
 private failed:{offset:number;append:boolean}|undefined;
 private view:SearchView=emptySearch;
 private nextOffset=0;
 constructor(private fetch:SearchHost['SearchLibrary'],private publish:(view:SearchView)=>void){}
 private update(view:SearchView){this.view=view;this.publish(view);}
 cancel(){this.generation++;clearTimeout(this.timer);this.current=undefined;this.failed=undefined;}
 search(profileID:string,request:SearchRequest){
  this.cancel();this.nextOffset=0;this.current={profileID,request:{...request,offset:0}};
  this.update({...emptySearch,status:'loading'});
  const generation=this.generation;
  this.timer=setTimeout(()=>void this.load(generation,0,false),180);
 }
 more(){if(this.current&&this.view.status==='ready'&&this.view.hasMore)void this.load(this.generation,this.nextOffset,true);}
 retry(){if(this.current&&this.view.status==='error'&&this.failed)void this.load(this.generation,this.failed.offset,this.failed.append);}
 private async load(generation:number,offset:number,append:boolean){
  const current=this.current;if(!current||generation!==this.generation)return;
  this.failed=undefined;
  this.update({...this.view,status:'loading',error:''});
  try{
   const page=await this.fetch(current.profileID,{...current.request,offset});
   if(generation!==this.generation)return;
   this.nextOffset=offset+page.hits.length;
   const hits=append?[...this.view.hits,...page.hits]:page.hits;
   // Reimports can move results between pages; never render duplicate IDs.
   const unique=[...new Map(hits.map(hit=>[hit.gameId,hit])).values()];
   this.update({...page,hits:unique,status:'ready',error:''});
  }catch(reason){
   if(generation!==this.generation)return;
   this.failed={offset,append};
   this.update({...this.view,status:'error',error:String(reason instanceof Error?reason.message:reason)});
  }
 }
}
