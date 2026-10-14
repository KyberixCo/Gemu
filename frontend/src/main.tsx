import {ActionHint} from './ActionHint';
import {useInputMethod} from './input-method';
import {navigationKey} from './keyboard-bindings';
import {FirstRunSetup} from './features/setup/FirstRunSetup';
import {GameSession,launchGame} from './GameSession';
import {toggleFullscreen} from './fullscreen';
import {ProfileGate} from './ProfileGate';
import {profileKey,type GemuProfile} from './profiles';
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { Search, Settings, Gamepad2, Keyboard, Mouse, Heart, X, LayoutGrid, Monitor, Maximize, Home, ChevronLeft, ChevronRight, Volume2, VolumeX, Sparkles, Boxes } from 'lucide-react';
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
import {subscribeGamepad} from './features/feedback/gamepad';
import {LocalizationProvider,useLocalization} from './localization';
import {VirtualInput} from './features/feedback/VirtualInput';
function read<T>(key:string, fallback:T):T { try { const value=JSON.parse(localStorage.getItem(profileKey(key))||'null'); return Array.isArray(value)&&value.every(x=>typeof x==='string') ? value as T : fallback; } catch { return fallback; } }
function App({profile,onManageProfiles}:{profile:GemuProfile;onManageProfiles:()=>void}) {
 const {language,setLanguage,t}=useLocalization();
 const input=useInputMethod();
 const [library,setLibrary]=useState<Snapshot>({version:1,sources:[],games:[]}),[demo,setDemo]=useState(true);
 const [providers,setProviders]=useState<SourceInfo[]>([]);
 const [sourcesStatus,setSourcesStatus]=useState<'loading'|'ready'|'error'>('loading'),[sourcesError,setSourcesError]=useState(''),[connectionAttempt,setConnectionAttempt]=useState(0);
 const games=useMemo(()=>demo?demoGames.map(game=>({...game,subtitle:t(game.subtitle),genre:t(game.genre),description:t(game.description)})):catalogFromLibrary(library,language),[demo,library,language]);
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
 const [notice,setNotice]=useState('');
 const [sessionBlocking,setSessionBlocking]=useState(false);
 const {preferences,updatePreferences,reduced,active}=useFeedback(setNotice);
 const [welcome,setWelcome]=useState<WelcomePhase>('ready');
 const welcomeRef=useRef<WelcomePhase>('ready');
 const welcomeTimer=useRef<ReturnType<typeof setTimeout>|undefined>(undefined);
 const [panelOpen,setPanelOpen]=useState(false);
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
   if(document.documentElement.dataset.gameplay==='true'||name==='menu')return;
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
   if(element instanceof HTMLTextAreaElement || element instanceof HTMLInputElement && element.type!=='range' || element instanceof HTMLSelectElement && !['ui-motion','language'].includes(element.id))return;
   if(element instanceof HTMLSelectElement&&element.id==='language'&&(name==='left'||name==='right')){setLanguage(name==='left'?'es':'en');return;}
   if(element instanceof HTMLInputElement&&element.type==='range'&&(element.id==='ambient-volume'||element.id==='ui-volume')&&(name==='left'||name==='right')){updatePreferences(element.id==='ambient-volume'?{ambientVolume:Math.max(0,Math.min(100,Number(element.value)+(name==='right'?5:-5)))}:{volume:Math.max(0,Math.min(100,Number(element.value)+(name==='right'?5:-5)))});return;}
   if(element instanceof HTMLSelectElement&&element.id==='ui-motion'&&(name==='left'||name==='right')){const values=['system','reduced','full'] as const;const index=values.indexOf(preferences.motion),next=Math.max(0,Math.min(2,index+(name==='right'?1:-1)));updatePreferences({motion:values[next]});return;}
   interfaceAudio.play('navigate',moveFocus(name as 'left',document.querySelector('.tv-editor')??document.querySelector('.artwork-flow')??(modal&&dialog.current?dialog.current:document)));
  };
  const onKey=(e:KeyboardEvent)=>{
   const key=navigationKey(e);
   if(welcomeRef.current!=='done'){
    if(key==='Escape'){e.preventDefault();finishWelcome();}
    else if((key==='Enter'||key===' ')&&welcomeRef.current==='playing'&&!e.repeat){e.preventDefault();finishWelcome();}
    else if((key==='Enter'||key===' ')&&e.repeat){e.preventDefault();}
    else if((key==='Enter'||key===' ')&&welcomeRef.current==='ready'){e.preventDefault();action('confirm');}
    else if(['ArrowLeft','ArrowRight','ArrowUp','ArrowDown'].includes(key)&&welcomeRef.current==='ready'){e.preventDefault();const root=document.querySelector('.welcome');const directions:Record<string,'left'|'right'|'up'|'down'>={ArrowLeft:'left',ArrowRight:'right',ArrowUp:'up',ArrowDown:'down'};if(root)moveFocus(directions[key],root);}
    else if(key==='Tab'){const buttons=[...document.querySelectorAll<HTMLElement>('.welcome button')];const index=buttons.indexOf(document.activeElement as HTMLElement);e.preventDefault();buttons[(index+(e.shiftKey?-1:1)+buttons.length)%buttons.length]?.focus();}
    return;
   }
   if((key==='Enter'||key===' ')&&e.repeat&&!(e.target as HTMLElement).matches('input,select,textarea')){e.preventDefault();return;}
   if(key==='/'&&!modal&&!(e.target as HTMLElement).matches('input,select,textarea')){e.preventDefault();document.querySelector<HTMLInputElement>('.search input')?.focus();interfaceAudio.play('navigate');return;}
   if(key==='Escape'){e.preventDefault();close();return;}
   if((key==='Enter'||key===' ')&&!(e.target as HTMLElement).matches('input,select,textarea')){e.preventDefault();if(!e.repeat)action('confirm');return;}
   if(key==='Tab') {
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
   if(dirs[key]) {e.preventDefault();action(dirs[key]);}
  };
  window.addEventListener('keydown',onKey);
  const unsubscribe=subscribeGamepad({action});
  return()=>{window.removeEventListener('keydown',onKey);unsubscribe();};
 },[modal,welcome,reduced,preferences,language]);
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
 const displayedDetail=detail?(games.find(game=>game.id===detail.id)??detail):null;
 useEffect(()=>{if(!active)return;const timer=setTimeout(()=>interfaceAudio.selectGame(activeGame),600);return()=>clearTimeout(timer);},[activeGame,active]);
 return <div className="shell" data-motion={reduced?"reduced":"full"} data-welcome={welcome} style={{'--art':activeGame.color} as React.CSSProperties}>
 <GameSession games={games} notify={setNotice} setBlocking={setSessionBlocking}/>
 <div className="console" inert={sessionBlocking||modal||welcome!=='done'}>
 {tab==='Gestión'?<div className="management-backdrop" aria-hidden="true"/>:<>
 <WorldBackdrop game={activeGame} reduced={reduced} enabled={preferences.ambientMotion&&welcome==='done'} active={active}/>
 <AmbientField enabled={preferences.ambientMotion&&welcome==='done'} active={active} reduced={reduced}/>
 <div className="world-shade"/>
 </>}
 <aside className="system-rail" aria-label={t('Sistema')}>
  <span className="brand" aria-label={"Gemu"}>{"g"}<span>.</span></span>
  <div className="rail-nav">{[{name:'Inicio',label:'Jugar',icon:Home},{name:'Biblioteca',label:'Biblioteca',icon:LayoutGrid},{name:'Favoritos',label:'Favoritos',icon:Heart},{name:'Gestión',label:'Gestión',icon:Boxes}].map(({name,label,icon:Icon})=><button data-nav key={name} className={'rail-button '+(tab===name?'active':'')} aria-label={t(label)} aria-pressed={tab===name} onClick={()=>changeTab(name)}><Icon size={23}/><span>{t(label)}</span></button>)}</div>
  <button data-nav className="rail-button rail-settings" aria-label={t('Ajustes')} onClick={openSettings}><Settings size={23}/><span>{t('Ajustes')}</span></button>
  <span className="kyberix-mark" title={"Kyberix"}>K</span>
 </aside>
 <header className="status-bar"><div className="location"><span className="status-line"/>{t(tab==='Inicio'?'JUGAR':tab).toUpperCase()}<button data-nav className="catalog-mode" onClick={()=>{setDemo(!demo);setDetail(null);setPanelOpen(false);}}>{t(demo?'CATÁLOGO DEMO · VER MIS JUEGOS':'MIS JUEGOS · VER DEMO')}</button></div><div className="status-right"><span className="input-device-icon" aria-label={t(input.kind==='gamepad'?'Mando conectado':'Teclado y ratón')}>{input.kind==='gamepad'?<Gamepad2 size={21}/>:input.kind==='mouse'?<Mouse size={21}/>:<Keyboard size={21}/>}</span><time>{time.toLocaleTimeString(language==='en'?'en':'es',{hour:'2-digit',minute:'2-digit',hour12:false})}</time><button data-nav className="avatar" aria-label={`${t('Perfiles')}: ${profile.name}`} title={profile.name} onClick={onManageProfiles}>{profile.name.slice(0,2).toLocaleUpperCase()}</button></div></header>
 <main className="console-main">
 {tab==='Gestión'?<Management initialArtworkGame={artworkGame} onArtworkOpened={()=>setArtworkGame(undefined)} onManageProfiles={onManageProfiles} profileName={profile.name} notify={setNotice} catalog={games} library={library} availableSources={providers} sourcesStatus={sourcesStatus} sourcesError={sourcesError} onRetrySources={()=>setConnectionAttempt(attempt=>attempt+1)} onLibraryChanged={acceptLibrary}/>:<>
 {games.length>0&&<section className="game-stage" aria-label={t('Juego seleccionado')}>
  <div className="game-info" key={activeGame.id}>
   <div className="game-meta"><GameIcon game={activeGame}/><span>{activeGame.provider}</span><span className="meta-dot"/><span>{activeGame.genre}</span>{favorites.includes(activeGame.id)&&<Heart size={16} fill="currentColor"/>}</div>
   <GameTitle game={activeGame} className={'game-title title-'+activeGame.id}/>
   <p className="game-tagline">{activeGame.subtitle}</p>
   <div className="stage-actions"><button data-nav className="primary" onClick={()=>openDetails(activeGame)}><ActionHint action="confirm"/> {t('Ver juego')}</button><button data-nav className="favorite" aria-label={t(favorites.includes(activeGame.id)?'Quitar de favoritos':'Añadir a favoritos')} aria-pressed={favorites.includes(activeGame.id)} onClick={()=>toggle(activeGame)}><Heart size={21} fill={favorites.includes(activeGame.id)?"currentColor":"none"}/></button><span className="installation-note">{t(activeGame.installed?'Instalado':demo?'Sin instalación · Demo':'Sin instalación')}</span></div>
  </div>
  <div className="world-index" aria-hidden="true"><span>{String(games.findIndex(g=>g.id===activeGame.id)+1).padStart(2,'0')}</span><span className="index-rule"/><span>GEMU<br/>WORLDS</span></div>
 </section>}
 <section className={'game-shelf '+(tab!=='Inicio'?'expanded':'')} aria-label={t('Biblioteca de juegos')}>
  <div className="shelf-toolbar"><div className="shelf-heading"><h2>{t(tab==='Inicio'?'Tus juegos':tab)}</h2><span>{String(visible.length).padStart(2,'0')}</span></div><div className="shelf-tools">{tab!=='Inicio'&&<div className="filters" aria-label={t("Fuente")}>{['Todos',...new Set(games.map(game=>game.provider))].map(s=><button data-nav key={s} className={source===s?"chosen":''} onClick={()=>changeSource(s)}>{s==='Todos'?t(s):s}</button>)}</div>}<label className="search"><Search size={18}/><input data-nav placeholder={t("Buscar")} value={query} onChange={e=>setQuery(e.target.value)} aria-label={t("Buscar juegos")}/><ActionHint action="search"/></label></div></div>
  <div className="game-strip">{visible.map(g=><button data-nav className={'game-card '+(activeGame.id===g.id?'selected':'')} key={g.id} onFocus={()=>setSelected(g)} onClick={()=>{setSelected(g);openDetails(g);}} aria-label={`${t('Ver')} ${g.title}`} aria-pressed={activeGame.id===g.id}><div className={'tile-art'+(g.cover?' has-grid':'')}><Scene game={g} miniature/><div className="tile-shade"/>{!g.cover&&<span dir="auto" className={'tile-title title-'+g.id}>{g.title}</span>}{favorites.includes(g.id)&&<Heart className="tile-heart" size={16} fill="currentColor"/>}</div><div className="tile-caption"><span dir="auto"><GameIcon game={g}/>{g.title}</span><span className="selection-dot"/></div></button>)}</div>
  {visible.length===0&&<div className="empty"><LayoutGrid size={28}/><div><h3>{t(!demo&&games.length===0?'Tu biblioteca está vacía':tab==='Favoritos'?'Aún no tienes favoritos':'No encontramos ese juego')}</h3><p>{t(!demo&&games.length===0?'Configura una fuente desde Gestión para incorporar tus juegos.':tab==='Favoritos'?'Guarda tus juegos con el corazón.':'Prueba otro nombre o cambia la fuente.')}</p></div><button data-nav onClick={()=>{changeTab('Biblioteca');setQuery('');setSource('Todos');}}>{t('Ver biblioteca')}</button></div>}
 </section>
 </>}
 </main>
 <footer className="control-bar"><div className="control-help"><span><ActionHint action="confirm"/> {t('Seleccionar')}</span><span><ActionHint action="back"/> {t('Volver')}</span><span className="keyboard-help"><ActionHint action="navigate"/> {t(tab==='Gestión'?'Navegar controles':'Cambiar juego')}</span></div><span className="position"><ChevronLeft size={13}/>{visible.length?String(visible.findIndex(g=>g.id===activeGame.id)+1).padStart(2,'0'):'00'} <span>/ {String(visible.length).padStart(2,'0')}</span><ChevronRight size={13}/></span><span className="control-device">{t(input.kind==='gamepad'?'Mando':input.kind==='mouse'?'Ratón':'Teclado')}</span></footer>
 </div>
 {panelPresent&&<div className="backdrop" data-open={modal&&panelReady} inert={!modal} aria-hidden={!modal}><div className="dialog" role="dialog" aria-modal={modal} aria-labelledby="dialog-title" ref={dialog}><button data-nav className="close icon-button" aria-label={t("Cerrar")} onClick={closePanel}><X/></button>{displayedDetail?<><p className="eyebrow">{displayedDetail.provider} / {displayedDetail.genre}</p><div id="dialog-title"><GameTitle game={displayedDetail} level={2}/></div><p>{displayedDetail.description}</p><div className="demo-note">{displayedDetail.installed?t("Juego instalado · Listo para jugar"):t("Título de demostración · Sin instalación")}</div>{displayedDetail.player&&<section><h3>{t("Perfil ·")} {displayedDetail.player.profile.name||t("Sin asociar")}</h3><p>{t(displayedDetail.player.message)}</p>{(displayedDetail.player.status==='ready'||displayedDetail.player.status==='cached')&&<><p>{displayedDetail.player.achievements.filter(a=>a.unlocked).length} / {displayedDetail.player.achievements.length} {t("logros desbloqueados")}</p><ul className="achievement-list">{displayedDetail.player.achievements.map(a=><li key={a.id}><strong>{a.unlocked?'✓ ':'○ '}{a.name}</strong><p>{a.description}</p></li>)}</ul></>}</section>}{displayedDetail.installed&&<button data-nav className="primary" onClick={()=>void launchGame(displayedDetail.id).then(()=>closePanel()).catch(error=>setNotice(String(error)))}>{t("Jugar ahora")}</button>}{displayedDetail.installed&&<button data-nav onClick={()=>{setArtworkGame(displayedDetail.id);setPanelOpen(false);setDetail(null);setTab('Gestión');}}>{t("Cambiar arte")}</button>}<button data-nav className="primary" onClick={()=>toggle(displayedDetail)}><Heart size={18}/>{favorites.includes(displayedDetail.id)?t("Quitar de favoritos"):t("Guardar en favoritos")}</button></>:<><p className="eyebrow">{t("SISTEMA")}</p><h2 id="dialog-title">{t("Ajustes")}</h2><p>{demo?t("Estás usando el catálogo de demostración."):t("Tus juegos, metadatos y perfil se guardan en este equipo.")}</p><div className="setting-row"><span>{input.kind==="gamepad"?<Gamepad2/>:<Keyboard/>} {t("Control")}</span><strong>{t(input.kind==="gamepad"?"Mando":"Teclado y ratón")}</strong></div><div className="setting-row"><span><Monitor/> {t("Fuentes")}</span><strong>{demo?t("Modo demostración"):`${library.sources.length} ${t('fuentes configuradas')}`}</strong></div><div className="feedback-settings">
 <div className="setting-row"><span>{preferences.sound?<Volume2/>:<VolumeX/>} {t('Sonidos de interfaz')}</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.sound} onClick={()=>updatePreferences({sound:!preferences.sound})}>{t(preferences.sound?'Activados':'Silenciados')}</button></div>
 <div className="setting-row volume-setting"><label htmlFor="ui-volume">{t('Volumen')}</label><div><input data-nav id="ui-volume" type="range" min="0" max="100" step="5" value={preferences.volume} onChange={e=>updatePreferences({volume:Number(e.target.value)})}/><output htmlFor="ui-volume">{preferences.volume}%</output></div></div>
 <div className="setting-row"><label htmlFor="ui-motion">{t('Movimiento')}</label><select data-nav id="ui-motion" value={preferences.motion} onChange={e=>updatePreferences({motion:e.target.value as 'system'|'reduced'|'full'})}><option value="system">{t('Según el sistema')}</option><option value="reduced">{t('Reducido')}</option><option value="full">{t('Completo')}</option></select></div>
 <div className="setting-row"><span>{t('Sonido de bienvenida')}</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.welcomeSound} onClick={()=>updatePreferences({welcomeSound:!preferences.welcomeSound})}>{t(preferences.welcomeSound?'Activado':'Desactivado')}</button></div>
 <div className="setting-row"><span>{t('Ambiente de fondo')}</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.ambient} onClick={()=>updatePreferences({ambient:!preferences.ambient})}>{t(preferences.ambient?'Activado':'Desactivado')}</button></div>
 <div className="setting-row volume-setting"><label htmlFor="ambient-volume">{t('Volumen ambiente')}</label><div><input data-nav id="ambient-volume" type="range" min="0" max="100" step="5" value={preferences.ambientVolume} onChange={e=>updatePreferences({ambientVolume:Number(e.target.value)})}/><output htmlFor="ambient-volume">{preferences.ambientVolume}%</output></div></div>
 <div className="setting-row"><span>{t('Iluminación dinámica')}</span><button data-nav className="setting-switch" role="switch" aria-checked={preferences.ambientMotion} onClick={()=>updatePreferences({ambientMotion:!preferences.ambientMotion})}>{t(preferences.ambientMotion?'Activado':'Estático')}</button></div>
 <div className="feedback-buttons"><button data-nav disabled={!preferences.sound||preferences.volume===0} onClick={()=>{void interfaceAudio.preview('confirm').then(played=>{if(!played)setNotice('El audio no está disponible. Puedes seguir navegando sin sonido.');});}}><Volume2 size={17}/> {t('Probar sonido')}</button><button data-nav onClick={replayWelcome}><Sparkles size={17}/> {t('Repetir bienvenida')}</button></div>
 <div className="setting-row"><label htmlFor="language">{t('Idioma de la aplicación')}</label><select data-nav id="language" value={language} onChange={e=>setLanguage(e.target.value as 'es'|'en')}><option value="es">{t('Español')}</option><option value="en">{t("English")}</option></select></div>
 </div><button data-nav className="primary" onClick={()=>void toggleFullscreen().catch(()=>setNotice('La pantalla completa no está disponible aquí.'))}><Maximize size={18}/> {t('Alternar pantalla completa')}</button></>}</div></div>}
 <Welcome phase={welcome} ambientMotion={preferences.ambientMotion} active={active} reduced={reduced} muted={(!preferences.sound||preferences.volume===0)&&(!preferences.ambient||preferences.ambientVolume===0)} onStart={startWelcome} onSkip={()=>finishWelcome()} onMute={()=>{const enable=(!preferences.sound||preferences.volume===0)&&(!preferences.ambient||preferences.ambientVolume===0);updatePreferences({sound:enable,ambient:enable,volume:preferences.volume||20,ambientVolume:preferences.ambientVolume||25});}}/>
 {notice&&<div role="status" className="toast">{t(notice)}</div>}
 </div>;
}
createRoot(document.getElementById('root')!).render(<React.StrictMode><LocalizationProvider><ProfileGate>{(profile,manage)=><FirstRunSetup key={profile.id}><App profile={profile} onManageProfiles={manage}/></FirstRunSetup>}</ProfileGate><VirtualInput/></LocalizationProvider></React.StrictMode>);
