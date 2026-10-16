import {expect,it,vi} from 'vitest';
import {catalogFromLibrary,contentHost,credentialHost,loadLibrarySession,libraryHost,artworkHost,type Snapshot} from './library';
it('preserves the real console for RetroArch without treating it as a genre or inventing achievements',()=>{
 const snapshot:Snapshot={version:1,sources:[{id:'retroarch',name:'RetroArch',root:'C:/RetroArch',profileId:''}],games:[{sourceId:'retroarch',game:{ID:'retroarch:one',Title:'Homebrew',Installations:[{Provider:'retroarch',ExternalID:'nes',Path:'D:/ROMs/Homebrew.nes',Installed:true}]},metadata:{platform:'Nintendo NES',description:'',genres:[],cover:'',background:'',origin:'local',status:'pending'},player:{profile:{id:'',name:'',origin:''},status:'pending',message:'',achievements:[]}}]};
 const game=catalogFromLibrary(snapshot)[0];expect(game.platform).toBe('Nintendo NES');expect(game.genre).toBe('Sin clasificar');expect(game.emulated).toBe(true);expect(game.installed).toBe(true);expect(game.player?.achievements).toEqual([]);
});
it('localizes generated library text while preserving names and fetched metadata',()=>{
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'steam:42',Title:'Biblioteca',Installations:[]},metadata:{description:'',genres:[],cover:'',background:'',origin:'steam',status:'pending'},player:{profile:{id:'player',name:'Perfil',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 const english=catalogFromLibrary(snapshot,'en')[0],spanish=catalogFromLibrary(snapshot,'es')[0];
 expect(english.title).toBe('Biblioteca');
 expect(english.subtitle).toBe('Profile · Perfil');
 expect(english.genre).toBe('Uncategorized');
 expect(english.description).toBe('Metadata not fetched yet.');
 expect(spanish.genre).toBe('Sin clasificar');
 snapshot.games[0].metadata.description='Descripción del proveedor';
 snapshot.games[0].metadata.genres=['Aventura'];
 expect(catalogFromLibrary(snapshot,'en')[0].description).toBe('Descripción del proveedor');
 expect(catalogFromLibrary(snapshot,'en')[0].genre).toBe('Aventura');
 expect(snapshot.games[0].game.Title).toBe('Biblioteca');
});
it('maps imported data without fabricating achievements or demo installations',()=>{
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'other',game:{ID:'other:42',Title:'Real game',Installations:[{Provider:'other',ExternalID:'42',Path:'C:/Game',Installed:true}]},metadata:{description:'Fetched',genres:['Adventure'],cover:'/library-assets/cover',background:'/library-assets/background',origin:'provider',status:'ready'},player:{profile:{id:'player',name:'Player',origin:'local'},status:'pending',message:'No key',achievements:[]}}]};
 const game=catalogFromLibrary(snapshot)[0];expect(game.installed).toBe(true);expect(game.provider).toBe('other');expect(game.player?.achievements).toEqual([]);expect(game.artworks).toEqual([{id:'background',type:'image',src:'/library-assets/background'}]);expect(game.cover).toBe('/library-assets/cover');
 const coverOnly=catalogFromLibrary({...snapshot,games:[{...snapshot.games[0],metadata:{...snapshot.games[0].metadata,background:''}}]})[0];
 expect(coverOnly.artworks).toBeUndefined();expect(coverOnly.cover).toBe('/library-assets/cover');
});

it('rotates only chosen backgrounds after an artwork edit and exposes the chosen logo',()=>{
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'one',Title:'One',Installations:[]},metadata:{description:'',genres:[],cover:'',background:'/primary',logo:'/logo',backgrounds:['/primary','/chosen'],origin:'steam',status:'ready',artworks:[{id:'chosen',kind:'background',url:'/chosen',provider:'steamgriddb',width:1920,height:1080},{id:'unselected',kind:'background',url:'/unselected',provider:'steamgriddb',width:1920,height:1080}]},player:{profile:{id:'',name:'',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 const game=catalogFromLibrary(snapshot)[0];expect(game.logo).toBe('/logo');expect(game.artworks?.map(a=>a.type==='image'?a.src:'')).toEqual(['/primary','/chosen']);
});

it('maps selected grids, logos and icons to their distinct presentation roles',()=>{
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'one',Title:'One',Installations:[]},metadata:{description:'',genres:[],cover:'/grid',background:'/hero',logo:'/logo',icon:'/icon',origin:'steamgriddb',status:'ready',artworks:[{id:'logo',kind:'logo',url:'/logo',provider:'steamgriddb',width:800,height:300},{id:'icon',kind:'icon',url:'/icon',provider:'steamgriddb',width:256,height:256}]},player:{profile:{id:'',name:'',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 const game=catalogFromLibrary(snapshot)[0];expect(game.cover).toBe('/grid');expect(game.logo).toBe('/logo');expect(game.icon).toBe('/icon');expect(game.artworks).toEqual([{id:'background',type:'image',src:'/hero'}]);
});
it('preserves the manual-logo selection state for title rendering',()=>{
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'one',Title:'MOUSE: P.I. FOR HIRE',Installations:[]},metadata:{description:'',genres:[],cover:'',background:'',logo:'/chosen-logo',origin:'steamgriddb',status:'ready',manualArt:['logo']},player:{profile:{id:'',name:'',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 expect(catalogFromLibrary(snapshot)[0].logoManuallySelected).toBe(true);
});
it('requires all artwork job methods before opening the flow on an old host',()=>{
 vi.stubGlobal('window',{go:{main:{App:{StartArtworkSearch:vi.fn()}}}});
 try{expect(()=>artworkHost()).toThrow('Reinicia Gemu');}finally{vi.unstubAllGlobals();}
});

it('keeps small covers and unselected covers out of scenery, but permits explicitly rotated covers',()=>{
 const candidate={id:'cover',kind:'cover' as const,url:'/cover',provider:'steamgriddb',width:1920,height:1080};
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'one',Title:'One',Installations:[]},metadata:{description:'',genres:[],cover:'/cover',background:'/background',backgrounds:['/background'],origin:'steam',status:'ready',artworks:[candidate,{...candidate,id:'other',url:'/other'}]},player:{profile:{id:'',name:'',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 const urls=()=>catalogFromLibrary(snapshot)[0].artworks?.map(a=>a.type==='image'?a.src:'');
 expect(urls()).toEqual(['/background','/cover']);
 snapshot.games[0].metadata.artworks![0]={...candidate,width:920,height:430};expect(urls()).toEqual(['/background']);
 snapshot.games[0].metadata.backgrounds!.push('/other');expect(urls()).toEqual(['/background','/other']);
 snapshot.games[0].metadata.cover='/background';expect(urls()).toEqual(['/background','/other']);
});
it('loads source adapters even when the saved library cannot be loaded',async()=>{
 const sources=vi.fn(),library=vi.fn(),error=vi.fn();
 await loadLibrarySession({GetSources:async()=>[{id:'steam',name:'Steam'}],GetLibrary:async()=>{throw new Error('Invalid saved catalog');}},{sources,library,error});
 expect(sources).toHaveBeenCalledWith([{id:'steam',name:'Steam'}]);expect(library).not.toHaveBeenCalled();expect(error).toHaveBeenCalledWith('library',expect.any(Error));
});
it('adds backgrounds and the eligible selected cover, deduplicating the primary image',()=>{
 const base={id:'primary',url:'/library-assets/primary.img',provider:'steamgriddb',width:1920,height:1080};
 const snapshot:Snapshot={version:1,sources:[],games:[{sourceId:'steam',game:{ID:'steam:42',Title:'Game',Installations:[]},metadata:{description:'',genres:[],cover:'/cover.img',background:base.url,origin:'steam-store',status:'ready',artworks:[{...base,kind:'background'},{...base,id:'second',kind:'background',url:'/library-assets/second.img'},{...base,id:'cover',kind:'cover',url:'/cover.img'},{...base,id:'logo',kind:'logo',url:'/logo.img'},{...base,id:'icon',kind:'icon',url:'/icon.img'}]},player:{profile:{id:'',name:'',origin:'local'},status:'pending',message:'',achievements:[]}}]};
 const game=catalogFromLibrary(snapshot)[0];
 expect(game.artworks).toEqual([{id:'background',type:'image',src:base.url},{id:'second',type:'image',src:'/library-assets/second.img'},{id:'scenery-cover',type:'image',src:'/cover.img'}]);
 expect(game.cover).toBe('/cover.img');
});
it('reports an outdated host without hiding the successfully loaded catalog',async()=>{
 const sources=vi.fn(),library=vi.fn(),error=vi.fn();const snapshot:Snapshot={version:1,sources:[],games:[]};
 await loadLibrarySession({GetSources:undefined as never,GetLibrary:async()=>snapshot},{sources,library,error});
 expect(sources).not.toHaveBeenCalled();expect(library).toHaveBeenCalledWith(snapshot);expect(error).toHaveBeenCalledWith('sources',expect.objectContaining({message:expect.stringContaining('dev.ps1')}));
});
it('explains an outdated provider bridge before any credential request is sent',()=>{
 const request=vi.fn();
 vi.stubGlobal('window',{go:{main:{App:{PrepareLibraryImport:request}}}});
 try{
  expect(()=>contentHost()).toThrow('Recargar la interfaz no actualiza el host');
  expect(()=>contentHost()).toThrow('dev.ps1');
  expect(request).not.toHaveBeenCalled();
 }finally{vi.unstubAllGlobals();}
});
it('accepts the updated native provider bridge',()=>{
 const check=vi.fn();vi.stubGlobal('window',{go:{main:{App:{CheckContentProvider:check}}}});
 try{expect(contentHost()).toBe(libraryHost());expect(contentHost().CheckContentProvider).toBe(check);}
 finally{vi.unstubAllGlobals();}
});
it('requires the complete secure credential bridge on an updated host',()=>{
 vi.stubGlobal('window',{go:{main:{App:{GetCredentialStatus:vi.fn(),SaveProviderCredentials:vi.fn()}}}});
 try{expect(()=>credentialHost()).toThrow('gestor seguro de credenciales');}
 finally{vi.unstubAllGlobals();}
 const methods={GetCredentialStatus:vi.fn(),SaveProviderCredentials:vi.fn(),DeleteProviderCredentials:vi.fn()};
 vi.stubGlobal('window',{go:{main:{App:methods}}});
 try{expect(credentialHost()).toBe(libraryHost());}
 finally{vi.unstubAllGlobals();}
});
