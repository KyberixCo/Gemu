import {FirstRunSetup} from './features/setup/FirstRunSetup';
import {ProfileGate} from './ProfileGate';
import {profileKey,type GemuProfile} from './profiles';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Search, Settings, Gamepad2, Heart, X, LayoutGrid, Monitor, Maximize, Home, ChevronLeft, ChevronRight, Volume2, VolumeX, Sparkles, Boxes } from 'lucide-react';
import { games as demoGames, filterGames, type Game } from './catalog';
import {libraryHost,loadLibrarySession,catalogFromLibrary,type Snapshot,type SourceInfo} from './library';
import { moveFocus } from './navigation';
import './style.css';
import {GameTitle,GameIcon} from './GameIdentity';
import { Scene } from './Scene';
import { Management } from './features/management/Management';
import { interfaceAudio } from './features/feedback/audio';
import { useFeedback } from './features/feedback/useFeedback';
import { Welcome, type WelcomePhase } from './features/feedback/Welcome';
import { WorldBackdrop } from './features/feedback/WorldBackdrop';
import { AmbientField } from './features/feedback/AmbientField';
import { GamepadRepeater } from './features/feedback/input';
function read<T>(key:string, fallback:T):T { try { const value=JSON.parse(localStorage.getItem(profileKey(key))||'null'); return Array.isArray(value)&&value.every(x=>typeof x==='string') ? value as T : fallback; } catch { return fallback; } }
function App({profile,onManageProfiles}:{profile:GemuProfile;onManageProfiles:()=>void}) {
 const [library,setLibrary]=useState<Snapshot>({version:1,sources:[],games:[]}),[demo,setDemo]=useState(true);
 const [providers,setProviders]=useState<SourceInfo[]>([]);
 const [sourcesStatus,setSourcesStatus]=useState<'loading'|'ready'|'error'>('loading'),[sourcesError,setSourcesError]=useState(''),[connectionAttempt,setConnectionAttempt]=useState(0);
 const games=useMemo(()=>demo?demoGames:catalogFromLibrary(library),[demo,library]);
 const acceptLibrary=(snapshot:Snapshot)=>{setLibrary(snapshot);setDemo(false);};
 useEffect(()=>{
  let cancelled=false;setSourcesStatus('loading');setSourcesError('');
  try{void loadLibrarySession(libraryHost(),{
   sources:providers=>{if(!cancelled){setProviders(providers);setSourcesStatus('ready');}},
   library:snapshot=>{if(!cancelled){setLibrary(snapshot);setDemo(false);}},
   error:(part,error)=>{if(cancelled)return;if(part==='sources'){setSourcesStatus('error');setSourcesError(error instanceof Error?error.message:String(error));}else setNotice('No se pudo cargar la biblioteca guardada.');},
  });}catch(error){setSourcesStatus('error');setSourcesError(error instanceof Error?error.message:String(error));}
  return()=>{cancelled=true;};
 },[connectionAttempt]);
 const [artworkGame,setArtworkGame]=useState<string>();
 const [tab,setTab]=useState('Inicio'), [query,setQuery]=useState(''), [source,setSource]=useState('Todos');
 const [favorites,setFavorites]=useState<string[]>(()=>read('gemu.favorites',[]));
 const [selected,setSelected]=useState<Game>(demoGames[0]), [detail,setDetail]=useState<Game|null>(null), [settings,setSettings]=useState(false);
 const [notice,setNotice]=useState(''), [connected,setConnected]=useState(false);
 const {preferences,updatePreferences,reduced,active}=useFeedback(setNotice);
 const [welcome,setWelcome]=useState<WelcomePhase>('ready');
 const welcomeRef=useRef<WelcomePhase>('ready');
 const welcomeTimer=useRef<ReturnType<typeof setTimeout>|undefined>(undefined);
 const [panelOpen,setPanelOpen]=useState(false);
 const gamepadRepeater=useRef(new GamepadRepeater());
 const finishWelcome=(skipSound=true)=>{clearTimeout(welcomeTimer.current);if(skipSound)interfaceAudio.stop();else interfaceAudio.cancelPending();welcomeRef.current='done';setWelcome('done');};
 const startWelcome=()=>{
  if(welcomeRef.current!=='ready')return;
  welcomeRef.current='playing';setWelcome('playing');
  void interfaceAudio.preview('welcome');
  welcomeTimer.current=setTimeout(()=>finishWelcome(false),reduced?1300:1850);
 };
 useEffect(()=>{startWelcome();return()=>{clearTimeout(welcomeTimer.current);welcomeRef.current='ready';};},[]);
 const ambientSession=welcome!=='ready';
 useEffect(()=>{interfaceAudio.setAmbientSession(ambientSession);return()=>interfaceAudio.setAmbientSession(false);},[ambientSession]);
 useEffect(()=>{if(welcome==='done')document.querySelector<HTMLElement>('.game-card.selected')?.focus({preventScroll:true});},[welcome]);
 const returnFocus=useRef<HTMLElement|null>(null), dialog=useRef<HTMLDivElement>(null);
 const [time,setTime]=useState(()=>new Date());
 useEffect(()=>{const timer=setInterval(()=>setTime(new Date()),30000);return()=>clearInterval(timer);},[]);
 const modal=panelOpen;
 const panelPresent=!!detail||settings;
 const [panelReady,setPanelReady]=useState(false);
 useEffect(()=>{
  if(!panelPresent){setPanelReady(false);return;}
  let second=0;const first=requestAnimationFrame(()=>{second=requestAnimationFrame(()=>setPanelReady(true));});
  return()=>{cancelAnimationFrame(first);cancelAnimationFrame(second);};
 },[panelPresent]);
 const openDetails=(game:Game)=>{setDetail(game);setSettings(false);setPanelOpen(true);interfaceAudio.play('open');};
 const openSettings=()=>{setSettings(true);setDetail(null);setPanelOpen(true);interfaceAudio.play('open');};
 const closePanel=()=>{if(!panelOpen)return;setPanelOpen(false);interfaceAudio.play('close');};
 useEffect(()=>{if(panelOpen||!panelPresent)return;const timer=setTimeout(()=>{setDetail(null);setSettings(false);},reduced?80:160);return()=>clearTimeout(timer);},[panelOpen,panelPresent,reduced]);
 
 useEffect(()=>{if(!notice)return;const t=setTimeout(()=>setNotice(''),4500);return()=>clearTimeout(t);},[notice]);
 useEffect(()=>{
  if(modal) { returnFocus.current=document.activeElement as HTMLElement; dialog.current?.querySelector<HTMLElement>('[data-nav]')?.focus(); }
  else if(welcomeRef.current==='done')returnFocus.current?.focus();
 },[modal]);
 useEffect(()=>{
  const close=()=>{if(welcomeRef.current==='done'&&!modal&&document.querySelector('.management')){const event=new Event('gemu-management-back',{cancelable:true});window.dispatchEvent(event);if(event.defaultPrevented)return;}if(welcomeRef.current!=='done'){finishWelcome();return;}closePanel();};
  const action=(name:string)=>{
   if(name==='back'){close();return;}
   if(welcomeRef.current!=='done'){
    if(name==='confirm'){if(welcomeRef.current==='ready')(document.activeElement as HTMLElement)?.click();else finishWelcome();}
    else if(welcomeRef.current==='ready'){const root=document.querySelector('.welcome');if(root)moveFocus(name as 'left',root);}
    return;
   }
   if(name==='confirm'){(document.activeElement as HTMLElement)?.click();return;}
   const element=document.activeElement;
   if(element instanceof HTMLInputElement&&element.type==='range'&&element.id!=='ambient-volume'&&element.id!=='ui-volume'&&(name==='left'||name==='right'))return;
   if(document.querySelector('.tv-editor')&&(name==='up'||name==='down')){interfaceAudio.play('navigate',moveFocus(name as 'up',document.querySelector('.tv-editor')!));return;}
   if(element instanceof HTMLTextAreaElement || element instanceof HTMLInputElement && element.type!=='range' || element instanceof HTMLSelectElement && element.id!=='ui-motion')return;
   if(element instanceof HTMLInputElement&&element.type==='range'&&(element.id==='ambient-volume'||element.id==='ui-volume')&&(name==='left'||name==='right')){updatePreferences(element.id==='ambient-volume'?{ambientVolume:Math.max(0,Math.min(100,Number(element.value)+(name==='right'?5:-5)))}:{volume:Math.max(0,Math.min(100,Number(element.value)+(name==='right'?5:-5)))});return;}
   if(element instanceof HTMLSelectElement&&element.id==='ui-motion'&&(name==='left'||name==='right')){const values=['system','reduced','full'] as const;const index=values.indexOf(preferences.motion),next=Math.max(0,Math.min(2,index+(name==='right'?1:-1)));updatePreferences({motion:values[next]});return;}
   interfaceAudio.play('navigate',moveFocus(name as 'left',document.querySelector('.tv-editor')??document.querySelector('.artwork-flow')??(modal&&dialog.current?dialog.current:document)));
  };
  const onKey=(e:KeyboardEvent)=>{
   if(welcomeRef.current!=='done'){
    if(e.key==='Escape'){e.preventDefault();finishWelcome();}
    else if((e.key==='Enter'||e.key===' ')&&welcomeRef.current==='playing'&&!e.repeat){e.preventDefault();finishWelcome();}
    else if((e.key==='Enter'||e.key===' ')&&e.repeat){e.preventDefault();}
    else if(['ArrowLeft','ArrowRight','ArrowUp','ArrowDown'].includes(e.key)&&welcomeRef.current==='ready'){e.preventDefault();const root=document.querySelector('.welcome');const directions:Record<string,'left'|'right'|'up'|'down'>={ArrowLeft:'left',ArrowRight:'right',ArrowUp:'up',ArrowDown:'down'};if(root)moveFocus(directions[e.key],root);}
    else if(e.key==='Tab'){const buttons=[...document.querySelectorAll<HTMLElement>('.welcome button')];const index=buttons.indexOf(document.activeElement as HTMLElement);e.preventDefault();buttons[(index+(e.shiftKey?-1:1)+buttons.length)%buttons.length]?.focus();}
    return;
   }
   if((e.key==='Enter'||e.key===' ')&&e.repeat&&!(e.target as HTMLElement).matches('input,select,textarea')){e.preventDefault();return;}
   if(e.key==='/'&&!modal&&!(e.target as HTMLElement).matches('input,select,textarea')){e.preventDefault();document.querySelector<HTMLInputElement>('.search input')?.focus();interfaceAudio.play('navigate');return;}
   if(e.key==='Escape'){close();return;}
   if(e.key==='Tab') {
    if(modal) {
     const els=[...dialog.current!.querySelectorAll<HTMLElement>('button,input,select,textarea')].filter(el=>!el.hasAttribute('disabled'));
     const index=els.indexOf(document.activeElement as HTMLElement);
     const next=els[(index+(e.shiftKey?-1:1)+els.length)%els.length];
     e.preventDefault();if(next&&next!==document.activeElement){next.focus();interfaceAudio.play('navigate');}
    } else {
     const before=document.activeElement;
     requestAnimationFrame(()=>{if(before!==document.activeElement)interfaceAudio.play('navigate');});
    }
    return;
   }
   if((e.target as HTMLElement).matches('input,select,textarea'))return;
   const dirs:Record<string,string>={ArrowLeft:'left',ArrowRight:'right',ArrowUp:'up',ArrowDown:'down'};
   if(dirs[e.key]) {e.preventDefault();action(dirs[e.key]);}
  };
  window.addEventListener('keydown',onKey);
  let frame=0;
  const poll=(now:number)=>{if(document.hidden||!document.hasFocus()){gamepadRepeater.current.reset();frame=requestAnimationFrame(poll);return;}const pad=navigator.getGamepads?.().find(p=>p?.connected);setConnected(Boolean(pad));
   const next=pad?(pad.buttons[0]?.pressed?'confirm':pad.buttons[1]?.pressed?'back':pad.buttons[12]?.pressed||pad.axes[1]<-.5?'up':pad.buttons[13]?.pressed||pad.axes[1]>.5?'down':pad.buttons[14]?.pressed||pad.axes[0]<-.5?'left':pad.buttons[15]?.pressed||pad.axes[0]>.5?'right':''):'';
   if(gamepadRepeater.current.accept(next,now))action(next);frame=requestAnimationFrame(poll);
  };frame=requestAnimationFrame(poll);return()=>{window.removeEventListener('keydown',onKey);cancelAnimationFrame(frame);};
 },[modal,welcome,reduced,preferences]);
 const toggle=(g:Game)=>{
  const exists=favorites.includes(g.id), next=exists?favorites.filter(id=>id!==g.id):[...favorites,g.id];
  setFavorites(next);
  try {localStorage.setItem(profileKey('gemu.favorites'),JSON.stringify(next));interfaceAudio.play(exists?'favorite-off':'favorite-on');}
  catch {setNotice('No se pudieron guardar los favoritos en este navegador.');interfaceAudio.play('error');}
 };
 const changeTab=(name:string)=>{if(tab!==name){setTab(name);interfaceAudio.play('section');}};
 const changeSource=(name:string)=>{if(source!==name){setSource(name);interfaceAudio.play('section');}};
 const replayWelcome=()=>{interfaceAudio.stop();setPanelOpen(false);setDetail(null);setSettings(false);welcomeRef.current='ready';setWelcome('ready');};
 const visible=filterGames(games,query,source,tab==='Favoritos',favorites);
 const activeGame=visible.find(g=>g.id===selected.id)??visible[0]??selected;
 useEffect(()=>{const timer=setTimeout(()=>interfaceAudio.selectGame(activeGame),600);return()=>clearTimeout(timer);},[activeGame]);
 return <div className="shell" data-motion={reduced?'reduced':'full'} data-welcome={welcome} style={{'--art':activeGame.color} as React.CSSProperties}>
 <div className="console" inert={modal||welcome!=='done'}>
 <WorldBackdrop game={activeGame} reduced={reduced} enabled={preferences.ambientMotion&&welcome==='done'} active={active}/>
 <AmbientField enabled={preferences.ambientMotion&&welcome==='done'} active={active} reduced={reduced}/>
 <div className="world-shade"/>
 <aside className="system-rail" aria-label="Sistema">
  <span className="brand" aria-label="Gemu">g<span>.</span></span>
  <div className="rail-nav">{[{name:'Inicio',label:'Jugar',icon:Home},{name:'Biblioteca',label:'Biblioteca',icon:LayoutGrid},{name:'Favoritos',label:'Favoritos',icon:Heart},{name:'Gestión',label:'Gestión',icon:Boxes}].map(({name,label,icon:Icon})=><button data-nav key={name} className={'rail-button '+(tab===name?'active':'')} aria-label={label} aria-pressed={tab===name} onClick={()=>changeTab(name)}><Icon size={23}/><span>{label}</span></button>)}</div>
  <button data-nav className="rail-button rail-settings" aria-label="Configuración" onClick={openSettings}><Settings size={23}/><span>Ajustes</span></button>
  <span className="kyberix-mark" title="Kyberix">K</span>
 </aside>
 <header className="status-bar"><div className="location"><span className="status-line"/>{tab==='Inicio'?'JUGAR':tab.toUpperCase()}<button data-nav className="catalog-mode" onClick={()=>{setDemo(!demo);setDetail(null);setPanelOpen(false);}}>{demo?'CATÁLOGO DEMO · VER MIS JUEGOS':'MIS JUEGOS · VER DEMO'}</button></div><div className="status-right"><Gamepad2 size={21} aria-label={connected?'Mando conectado':'Teclado activo'}/><time>{time.toLocaleTimeString('es',{hour:'2-digit',minute:'2-digit',hour12:false})}</time><button data-nav className="avatar" aria-label={`Perfiles: ${profile.name}`} title={profile.name} onClick={onManageProfiles}>{profile.name.slice(0,2).toLocaleUpperCase()}</button></div></header>
 <main className="console-main">
 {tab==='Gestión'?<Management initialArtworkGame={artworkGame} onArtworkOpened={()=>setArtworkGame(undefined)} onManageProfiles={onManageProfiles} profileName={profile.name} notify={setNotice} catalog={games} library={library} availableSources={providers} sourcesStatus={sourcesStatus} sourcesError={sourcesError} onRetrySources={()=>setConnectionAttempt(attempt=>attempt+1)} onLibraryChanged={acceptLibrary}/>:<>
 {games.length>0&&<section className="game-stage" aria-label="Juego seleccionado">
  <div className="game-info" key={activeGame.id}>
   <div className="game-meta"><GameIcon game={activeGame}/><span>{activeGame.provider}</span><span className="meta-dot"/><span>{activeGame.genre}</span>{favorites.includes(activeGame.id)&&<Heart size={16} fill="currentColor"/>}</div>
   <GameTitle game={activeGame} className={'game-title title-'+activeGame.id}/>
   <p className="game-tagline">{activeGame.subtitle}</p>
   <div className="stage-actions"><button data-nav className="primary" onClick={()=>openDetails(activeGame)}><span className="pad-key">A</span> Ver juego</button><button data-nav className="favorite" aria-label={favorites.includes(activeGame.id)?'Quitar de favoritos':'Añadir a favoritos'} aria-pressed={favorites.includes(activeGame.id)} onClick={()=>toggle(activeGame)}><Heart size={21} fill={favorites.includes(activeGame.id)?'currentColor':'none'}/></button><span className="installation-note">{activeGame.installed?'Instalado':demo?'Sin instalación · Demo':'Sin instalación'}</span></div>
  </div>
  <div className="world-index" aria-hidden="true"><span>{String(games.findIndex(g=>g.id===activeGame.id)+1).padStart(2,'0')}</span><span className="index-rule"/><span>GEMU<br/>WORLDS</span></div>
 </section>}
 <section className={'game-shelf '+(tab!=='Inicio'?'expanded':'')} aria-label="Biblioteca de juegos">
  <div className="shelf-toolbar"><div className="shelf-heading"><h2>{tab==='Inicio'?'Tus juegos':tab}</h2><span>{String(visible.length).padStart(2,'0')}</span></div><div className="shelf-tools">{tab!=='Inicio'&&<div className="filters" aria-label="Fuente">{['Todos',...new Set(games.map(game=>game.provider))].map(s=><button data-nav key={s} className={source===s?'chosen':''} onClick={()=>changeSource(s)}>{s}</button>)}</div>}<label className="search"><Search size={18}/><input data-nav placeholder="Buscar" value={query} onChange={e=>setQuery(e.target.value)} aria-label="Buscar juegos"/><kbd>/</kbd></label></div></div>
  <div className="game-strip">{visible.map(g=><button data-nav className={'game-card '+(activeGame.id===g.id?'selected':'')} key={g.id} onFocus={()=>setSelected(g)} onClick={()=>{setSelected(g);openDetails(g);}} aria-label={`Ver ${g.title}`} aria-pressed={activeGame.id===g.id}><div className={'tile-art'+(g.cover?' has-grid':'')}><Scene game={g} miniature/><div className="tile-shade"/>{!g.cover&&<span className={'tile-title title-'+g.id}>{g.title}</span>}{favorites.includes(g.id)&&<Heart className="tile-heart" size={16} fill="currentColor"/>}</div><div className="tile-caption"><span><GameIcon game={g}/>{g.title}</span><span className="selection-dot"/></div></button>)}</div>
  {visible.length===0&&<div className="empty"><LayoutGrid size={28}/><div><h3>{!demo&&games.length===0?'Tu biblioteca está vacía':tab==='Favoritos'?'Aún no tienes favoritos':'No encontramos ese juego'}</h3><p>{!demo&&games.length===0?'Configura una fuente desde Gestión para incorporar tus juegos.':tab==='Favoritos'?'Guarda tus juegos con el corazón.':'Prueba otro nombre o cambia la fuente.'}</p></div><button data-nav onClick={()=>{changeTab('Biblioteca');setQuery('');setSource('Todos');}}>Ver biblioteca</button></div>}
 </section>
 </>}
 </main>
 <footer className="control-bar"><div className="control-help"><span><span className="pad-key">A</span> Seleccionar</span><span><span className="pad-key">B</span> Volver</span><span className="keyboard-help"><kbd>← →</kbd> {tab==='Gestión'?'Navegar controles':'Cambiar juego'}</span></div><span className="position"><ChevronLeft size={13}/>{visible.length?String(visible.findIndex(g=>g.id===activeGame.id)+1).padStart(2,'0'):'00'} <span>/ {String(visible.length).padStart(2,'0')}</span><ChevronRight size={13}/></span><span className="control-device">{connected?'MANDO CONECTADO':'TECLADO · ENTER / ESC'}</span></footer>
 </div>
 {panelPresent&&<div className="backdrop" data-open={modal&&panelReady} inert={!modal} aria-hidden={!modal}><div className="dialog" role="dialog" aria-modal={modal} aria-labelledby="dialog-title" ref={dialog}><button data-nav className="close icon-button" aria-label="Cerrar" onClick={closePanel}><X/></button>{detail?<><p className="eyebrow">{detail.provider} / {detail.genre}</p><div id="dialog-title"><GameTitle game={detail} level={2}/></div><p>{detail.description}</p><div className="demo-note">{detail.installed?'Juego instalado · Lanzamiento pendiente de integrar':'Título de demostración · Sin instalación'}</div>{detail.player&&<section><h3>Perfil · {detail.player.profile.name||'Sin asociar'}</h3><p>{detail.player.message}</p>{(detail.player.status==='ready'||detail.player.status==='cached')&&<><p>{detail.player.achievements.filter(a=>a.unlocked).length} / {detail.player.achievements.length} logros desbloqueados</p><ul className="achievement-list">{detail.player.achievements.map(a=><li key={a.id}><strong>{a.unlocked?'✓ ':'○ '}{a.name}</strong><p>{a.description}</p></li>)}</ul></>}</section>}{detail.installed&&<button data-nav onClick={()=>{setArtworkGame(detail.id);setPanelOpen(false);setDetail(null);setTab('Gestión');}}>Cambiar arte</button>}<button data-nav className="primary" onClick={()=>toggle(detail)}><Heart size={18}/>{favorites.includes(detail.id)?'Quitar de favoritos':'Guardar en favoritos'}</button></>:<><p className="eyebrow">SISTEMA</p><h2 id="dialog-title">Ajustes</h2><p>{demo?'Estás usando el catálogo de demostración.':'Tus juegos, metadatos y perfil se guardan en este equipo.'}</p><div className="setting-row"><span><Gamepad2/> Control</span><strong>{connected?'Mando conectado':'Teclado y ratón'}</strong></div><div className="setting-row"><span><Monitor/> Fuentes</span><strong>{demo?'Modo demostración':`${library.sources.length} fuentes configuradas`}</strong></div><div className="feedback-settings">
 <div className="setting-row"><span>{preferences.sound?<Volume2/>:<VolumeX/>} Sonidos de interfaz</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.sound} onClick={()=>updatePreferences({sound:!preferences.sound})}>{preferences.sound?'Activados':'Silenciados'}</button></div>
 <div className="setting-row volume-setting"><label htmlFor="ui-volume">Volumen</label><div><input data-nav id="ui-volume" type="range" min="0" max="100" step="5" value={preferences.volume} onChange={e=>updatePreferences({volume:Number(e.target.value)})}/><output htmlFor="ui-volume">{preferences.volume}%</output></div></div>
 <div className="setting-row"><label htmlFor="ui-motion">Movimiento</label><select data-nav id="ui-motion" value={preferences.motion} onChange={e=>updatePreferences({motion:e.target.value as 'system'|'reduced'|'full'})}><option value="system">Según el sistema</option><option value="reduced">Reducido</option><option value="full">Completo</option></select></div>
 <div className="setting-row"><span>Sonido de bienvenida</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.welcomeSound} onClick={()=>updatePreferences({welcomeSound:!preferences.welcomeSound})}>{preferences.welcomeSound?'Activado':'Desactivado'}</button></div>
 <div className="setting-row"><span>Ambiente de fondo</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.ambient} onClick={()=>updatePreferences({ambient:!preferences.ambient})}>{preferences.ambient?'Activado':'Desactivado'}</button></div>
 <div className="setting-row volume-setting"><label htmlFor="ambient-volume">Volumen ambiente</label><div><input data-nav id="ambient-volume" type="range" min="0" max="100" step="5" value={preferences.ambientVolume} onChange={e=>updatePreferences({ambientVolume:Number(e.target.value)})}/><output htmlFor="ambient-volume">{preferences.ambientVolume}%</output></div></div>
 <div className="setting-row"><span>Iluminación dinámica</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.ambientMotion} onClick={()=>updatePreferences({ambientMotion:!preferences.ambientMotion})}>{preferences.ambientMotion?'Activado':'Estático'}</button></div>
 <div className="feedback-buttons"><button data-nav disabled={!preferences.sound||preferences.volume===0} onClick={()=>{void interfaceAudio.preview('confirm').then(played=>{if(!played)setNotice('El audio no está disponible. Puedes seguir navegando sin sonido.');});}}><Volume2 size={17}/> Probar sonido</button><button data-nav onClick={replayWelcome}><Sparkles size={17}/> Repetir bienvenida</button></div>
 </div><button data-nav className="primary" onClick={()=>{if(document.fullscreenElement)document.exitFullscreen().catch(()=>setNotice('No se pudo salir de pantalla completa.'));else document.documentElement.requestFullscreen().catch(()=>setNotice('La pantalla completa no está disponible aquí.'));}}><Maximize size={18}/> Alternar pantalla completa</button></>}</div></div>}
 <Welcome phase={welcome} ambientMotion={preferences.ambientMotion} active={active} reduced={reduced} muted={(!preferences.sound||preferences.volume===0)&&(!preferences.ambient||preferences.ambientVolume===0)} onStart={startWelcome} onSkip={()=>finishWelcome()} onMute={()=>{const enable=(!preferences.sound||preferences.volume===0)&&(!preferences.ambient||preferences.ambientVolume===0);updatePreferences({sound:enable,ambient:enable,volume:preferences.volume||20,ambientVolume:preferences.ambientVolume||25});}}/>
 {notice&&<div role="status" className="toast">{notice}</div>}
 </div>;
}
createRoot(document.getElementById('root')!).render(<React.StrictMode><ProfileGate>{(profile,manage)=><FirstRunSetup key={profile.id}><App profile={profile} onManageProfiles={manage}/></FirstRunSetup>}</ProfileGate></React.StrictMode>);
