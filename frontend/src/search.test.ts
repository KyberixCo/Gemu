import {afterEach,expect,it,vi} from 'vitest';
import {LibrarySearch,type SearchPage,type SearchRequest,type SearchView} from './search';

const request:SearchRequest={query:'orbit',sourceId:'steam',favoritesOnly:false,limit:2,offset:0};
const page=(ids:string[],total=ids.length,hasMore=false):SearchPage=>({hits:ids.map(gameId=>({gameId,title:gameId,kind:'game',sourceId:'steam',score:-1})),total,hasMore});
function deferred<T>(){let resolve!:(value:T)=>void;let reject!:(error:Error)=>void;const promise=new Promise<T>((yes,no)=>{resolve=yes;reject=no;});return {promise,resolve,reject};}
afterEach(()=>vi.useRealTimers());

it('debounces typing and sends source and favorite filters to the host',async()=>{
 vi.useFakeTimers();const fetch=vi.fn(async()=>page(['final']));const update=vi.fn();const search=new LibrarySearch(fetch,update);
 search.search('a',{...request,query:'o'});search.search('a',{...request,query:'orbit',favoritesOnly:true});
 await vi.advanceTimersByTimeAsync(179);expect(fetch).not.toHaveBeenCalled();
 await vi.advanceTimersByTimeAsync(1);
 expect(fetch).toHaveBeenCalledExactlyOnceWith('a',{...request,favoritesOnly:true});
 expect(update.mock.lastCall?.[0]).toMatchObject({status:'ready',hits:[{gameId:'final'}]});
});

it('ignores late results and errors from previous queries and players',async()=>{
 vi.useFakeTimers();const old=deferred<SearchPage>(),lateError=deferred<SearchPage>();
 const fetch=vi.fn().mockReturnValueOnce(old.promise).mockReturnValueOnce(lateError.promise).mockResolvedValue(page(['b-result']));
 const update=vi.fn();const search=new LibrarySearch(fetch,update);
 search.search('a',request);await vi.advanceTimersByTimeAsync(180);
 search.search('a',{...request,query:'second'});await vi.advanceTimersByTimeAsync(180);
 search.search('b',request);await vi.advanceTimersByTimeAsync(180);
 const calls=update.mock.calls.length;
 old.resolve(page(['a-secret']));lateError.reject(new Error('old failure'));await vi.advanceTimersByTimeAsync(0);
 expect(update).toHaveBeenCalledTimes(calls);expect(update.mock.lastCall?.[0].hits[0].gameId).toBe('b-result');
});

it('paginates, blocks repeated requests, and retries a failed page without losing earlier results',async()=>{
 vi.useFakeTimers();const pending=deferred<SearchPage>();
 const fetch=vi.fn().mockResolvedValueOnce(page(['one','two'],3,true)).mockReturnValueOnce(pending.promise).mockResolvedValueOnce(page(['three'],3));
 let view:SearchView|undefined;const search=new LibrarySearch(fetch,next=>{view=next;});
 search.search('a',request);await vi.advanceTimersByTimeAsync(180);
 search.more();search.more();expect(fetch).toHaveBeenCalledTimes(2);expect(fetch.mock.calls[1][1].offset).toBe(2);
 pending.reject(new Error('disk unavailable'));await vi.advanceTimersByTimeAsync(0);
 expect(view).toMatchObject({status:'error',hits:[{gameId:'one'},{gameId:'two'}]});
 search.retry();await vi.advanceTimersByTimeAsync(0);
 expect(fetch.mock.calls[2][1].offset).toBe(2);expect(view).toMatchObject({status:'ready',total:3,hasMore:false});
 expect(view?.hits.map(hit=>hit.gameId)).toEqual(['one','two','three']);
});

it('cancels timers and ignores in-flight completions after leaving the search',async()=>{
 vi.useFakeTimers();const pending=deferred<SearchPage>();const fetch=vi.fn(()=>pending.promise),update=vi.fn();const search=new LibrarySearch(fetch,update);
 search.search('a',request);search.cancel();await vi.advanceTimersByTimeAsync(180);expect(fetch).not.toHaveBeenCalled();
 search.search('a',request);await vi.advanceTimersByTimeAsync(180);search.cancel();const count=update.mock.calls.length;
 pending.resolve(page(['late']));await vi.advanceTimersByTimeAsync(0);expect(update).toHaveBeenCalledTimes(count);
});

it('keeps pagination offsets even when a reimport moves a result between pages',async()=>{
 vi.useFakeTimers();const fetch=vi.fn().mockResolvedValueOnce(page(['one','two'],5,true)).mockResolvedValueOnce(page(['two','three'],5,true)).mockResolvedValueOnce(page(['four'],5));
 let view:SearchView|undefined;const search=new LibrarySearch(fetch,next=>{view=next;});
 search.search('a',request);await vi.advanceTimersByTimeAsync(180);search.more();await vi.advanceTimersByTimeAsync(0);search.more();await vi.advanceTimersByTimeAsync(0);
 expect(fetch.mock.calls[2][1].offset).toBe(4);expect(view?.hits.map(hit=>hit.gameId)).toEqual(['one','two','three','four']);
});
