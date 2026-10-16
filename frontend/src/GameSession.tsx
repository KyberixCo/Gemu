import {useEffect,useRef,useState,type Ref} from 'react';
import {ArrowLeft,ArrowRight,Check,Gamepad2,LayoutGrid,MousePointer2,Play,Power,TriangleAlert} from 'lucide-react';
import {subscribeGamepad} from './features/feedback/gamepad';
import {interfaceAudio} from './features/feedback/audio';
import {moveFocus} from './navigation';
import {useLocalization} from './localization';
import {ActionHint} from './ActionHint';
import {Scene} from './Scene';
import {GameTitle} from './GameIdentity';
import type {Game} from './catalog';
import './game-session.css';
import {setAssisting,setGameplay} from './activity';
import {useInputMethod} from './input-method';
import {navigationKey} from './keyboard-bindings';

export type Session={gameId:string;title:string;state:string;menu:boolean;overlay?:boolean;background:boolean;message:string};
type Host={LaunchGame:(id:string)=>Promise<Session>;GetGameSession:()=>Promise<Session>;OpenGameMenu:()=>Promise<Session>;ReturnToGemu:()=>Promise<Session>;ResumeGame:()=>Promise<void>;CloseGame:()=>Promise<void>;OpenInputAssistant?:()=>Promise<void>;SetGameOverlayLanguage?:(language:string)=>Promise<void>};
function host(){const app=(window as Window&{go?:{main?:{App?:Host}}}).go?.main?.App;if(!app?.LaunchGame)throw new Error('Reinicia Gemu para usar el lanzamiento de juegos.');return app;}
export async function launchGame(id:string){setGameplay(true);try{return await host().LaunchGame(id);}catch(error){setGameplay(false);throw error;}}
const isActive=(session:Session|null)=>!!session&&['requested','running','closing'].includes(session.state);

type MenuProps={session:Session;game?:Game;confirm:boolean;busy:boolean;error:string;menuRef?:Ref<HTMLDivElement>;onResume:()=>void;onHome:()=>void;onClose:()=>void;onCancel:()=>void;onConfirm:()=>void;onAssist?:()=>void};

export function SessionMenu({session,game,confirm,busy,error,menuRef,onResume,onHome,onClose,onCancel,onConfirm,onAssist}:MenuProps){
 const {t}=useLocalization();
 const input=useInputMethod();
 const status=t(session.state==='requested'?'Iniciando…':session.state==='closing'?'Cerrando…':'En ejecución');
 return <div className="session-backdrop">
  <div className="session-menu" data-input={input.kind} ref={menuRef} role={confirm?'alertdialog':'dialog'} aria-modal="true" aria-labelledby={confirm?'session-confirm-title':'session-title'} aria-describedby={confirm?'session-warning':'session-description'} aria-busy={busy}>
   <div className="session-art" aria-hidden="true">{game&&<Scene game={game}/>}</div>
   <header className="session-header">
    <span className="session-brand" aria-label="Gemu">g<span>.</span></span>
    <span className="session-section">{t('Partida actual')}</span>
    <span className="session-status" data-state={session.state}><span aria-hidden="true"/>{status}</span>
   </header>
   <div className="session-content">
    <section className="session-controls" aria-label={t('Menú del juego')}>
     <p className="session-eyebrow">GEMU / {t('Menú del juego')}</p>
     <h2 className="session-menu-heading" id={confirm?'session-confirm-title':undefined}>{t(confirm?'¿Cerrar el juego?':'Tu partida')}</h2>
     {confirm?<>
      <p className="session-warning" id="session-warning"><TriangleAlert size={22} aria-hidden="true"/>{t('El progreso sin guardar puede perderse.')}</p>
      <div className="session-actions">
       <button data-nav data-session-action="cancel" className="session-action session-primary" disabled={busy} onClick={onCancel}><ArrowLeft aria-hidden="true"/><span><strong>{t('Cancelar')}</strong><small>{t('Volver a las opciones de la partida.')}</small></span></button>
       <button data-nav className="session-action session-danger" disabled={busy||session.state!=='running'} onClick={onConfirm}><Power aria-hidden="true"/><span><strong>{t('Confirmar cierre')}</strong><small>{t('Solicitar el cierre del juego.')}</small></span><Check className="session-action-end" aria-hidden="true"/></button>
      </div>
     </>:<div className="session-actions">
      <button data-nav className="session-action session-primary" disabled={busy||session.state==='requested'} onClick={onResume}><Play aria-hidden="true"/><span><strong>{t('Continuar jugando')}</strong><small>{t('Regresar directamente al juego.')}</small></span><ArrowRight className="session-action-end" aria-hidden="true"/></button>
      <button data-nav className="session-action" disabled={busy} onClick={onHome}><LayoutGrid aria-hidden="true"/><span><strong>{t('Volver a Gemu')}</strong><small>{t('Explorar tu biblioteca sin cerrar el juego.')}</small></span></button>
      {onAssist&&<button data-nav className="session-action" disabled={busy} onClick={onAssist}><MousePointer2 aria-hidden="true"/><span><strong>{t('Asistencia de escritorio')}</strong><small>{t('Usar ratón y teclado en el juego o su ventana de inicio.')}</small></span></button>}
      <button data-nav data-session-action="close" className="session-action session-exit" disabled={busy||session.state!=='running'} onClick={onClose}><Power aria-hidden="true"/><span><strong>{t('Cerrar juego')}</strong><small>{t('Terminar la partida.')}</small></span></button>
     </div>}
     <div className="session-feedback" aria-live="polite">{busy&&<p>{t('Procesando…')}</p>}{session.message&&<p role="status">{t(session.message)}</p>}{error&&<p className="session-error" role="alert">{t(error)}</p>}</div>
    </section>
    <section className="session-game">
     <p className="session-eyebrow"><Gamepad2 size={18} aria-hidden="true"/>{t('Partida actual')}{game?.provider&&<span> / {game.provider}</span>}</p>
     {game?<GameTitle game={{...game,title:session.title}} level={2} className="session-game-title"/>:<h2 className="session-game-title" dir="auto">{session.title}</h2>}
     <span id="session-title" className="session-sr-only">{session.title}</span>
     <p id="session-description" className="session-game-note">{t('La partida sigue ejecutándose mientras este menú está abierto.')}</p>
    </section>
   </div>
   <footer className="session-footer">
    <div className="session-control-hints"><span><ActionHint action="navigate" axis="vertical"/>{t('Navegar')}</span><span><ActionHint action="confirm"/>{t('Seleccionar')}</span><span><ActionHint action="back"/>{t(confirm?'Cancelar':session.state==='requested'?'Volver a Gemu':'Continuar jugando')}</span></div>
    <span className="session-shortcut"><kbd>F10</kbd>{t('Menú del juego')}{input.kind==='gamepad'&&input.family==='xbox'&&<span> · Home / View + Menu</span>}</span>
   </footer>
  </div>
 </div>;
}

export function GameSession({games,notify,setBlocking}:{games:Game[];notify:(text:string)=>void;setBlocking:(value:boolean)=>void}){
 const {t,language}=useLocalization();
 useEffect(()=>{try{void host().SetGameOverlayLanguage?.(language)?.catch(()=>{});}catch{}},[language]);
 const [session,setSession]=useState<Session|null>(null),[confirm,setConfirm]=useState(false),[busy,setBusy]=useState(false),[error,setError]=useState('');
 const current=useRef(session),root=useRef<HTMLDivElement>(null),priorFocus=useRef<HTMLElement|null>(null),pending=useRef(false),wasOpen=useRef(false),wasConfirm=useRef(false);
 current.current=session;
 const active=isActive(session);
 useEffect(()=>{setGameplay(active&&(!session?.menu||!!session?.overlay)&&!session?.background);},[active,session?.menu,session?.overlay,session?.background]);
 useEffect(()=>()=>setGameplay(false),[]);
 useEffect(()=>{setBlocking(active&&!session?.background);return()=>setBlocking(false);},[active,session?.background,setBlocking]);
 const run=async(action:()=>Promise<unknown>)=>{
  if(pending.current)return;
  pending.current=true;setBusy(true);setError('');
  try{await action();setSession(await host().GetGameSession());}
  catch(error){setError(String(error));notify(String(error));}
  finally{pending.current=false;setBusy(false);}
 };
 useEffect(()=>{
  let disposed=false,version=0,timer:ReturnType<typeof setTimeout>|undefined;
  const accept=(next:Session)=>{if(!disposed)setSession(previous=>previous&&JSON.stringify(previous)===JSON.stringify(next)?previous:next);};
  const runtime=(window as Window&{runtime?:{EventsOn?:(name:string,callback:(next:Session)=>void)=>()=>void}}).runtime;
  const unsubscribe=runtime?.EventsOn?.('gemu:session',next=>{version++;accept(next);});
  const assistRuntime=(window as Window&{runtime?:{EventsOn?:(name:string,callback:(active:boolean)=>void)=>()=>void}}).runtime;
  const stopAssist=assistRuntime?.EventsOn?.('gemu:assist',active=>setAssisting(active));
  let polling=false;
  const poll=async()=>{if(disposed||polling)return;polling=true;clearTimeout(timer);const before=version;try{const next=await host().GetGameSession();if(version===before)accept(next);}catch{/* Desktop bridge is optional in browser previews. */}finally{polling=false;}if(!disposed&&!unsubscribe)timer=setTimeout(()=>void poll(),2000);};
  void poll();const refresh=()=>void poll();window.addEventListener('focus',refresh);
  return()=>{disposed=true;clearTimeout(timer);unsubscribe?.();stopAssist?.();setAssisting(false);window.removeEventListener('focus',refresh);};
 },[]);
 useEffect(()=>{document.documentElement.dataset.gameplay=active&&(!session?.menu||!!session?.overlay)&&!session?.background?'true':'false';return()=>{delete document.documentElement.dataset.gameplay;};},[active,session?.menu,session?.overlay,session?.background]);
 useEffect(()=>{
  if(session?.menu){
   if(!wasOpen.current){priorFocus.current=document.activeElement as HTMLElement;setConfirm(false);setError('');interfaceAudio.play('open');}
   const selector=confirm?'[data-session-action="cancel"]':wasConfirm.current?'[data-session-action="close"]':'button:not(:disabled)';
   const target=root.current?.querySelector<HTMLButtonElement>(selector);
   if(target&&!target.disabled)target.focus({preventScroll:true});
   else root.current?.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus({preventScroll:true});
  }else if(wasOpen.current){setConfirm(false);priorFocus.current?.focus({preventScroll:true});}
  wasOpen.current=!!session?.menu;wasConfirm.current=confirm;
 },[session?.menu,confirm]);
 useEffect(()=>{if(!busy&&session?.menu&&!root.current?.contains(document.activeElement))root.current?.querySelector<HTMLButtonElement>('button:not(:disabled)')?.focus({preventScroll:true});},[busy,session?.menu]);
 useEffect(()=>{
  const resume=()=>void run(()=>current.current?.state==='requested'?host().ReturnToGemu():host().ResumeGame());
  const action=(name:string)=>{
   const value=current.current;if(!isActive(value)||!value||pending.current)return;
   if(name==='menu'){if(!value.menu)void run(()=>host().OpenGameMenu());return;}
   if(!value.menu)return;
   if(name==='back'){if(confirm)setConfirm(false);else resume();}
   else if(name==='confirm'){if(root.current?.contains(document.activeElement))(document.activeElement as HTMLElement)?.click();}
   else if(root.current)interfaceAudio.play('navigate',moveFocus(name as 'left',root.current));
  };
  const unsubscribe=subscribeGamepad({capture:()=>{const value=current.current;return isActive(value)&&!value?.background;},action});
  const key=(event:KeyboardEvent)=>{
   const value=current.current;if(!isActive(value)||!value)return;
   if(event.key==='F10'){event.preventDefault();event.stopImmediatePropagation();return;}
   if(!value.menu)return;
   event.stopImmediatePropagation();const pressed=navigationKey(event);
   if(pressed==='Escape'){event.preventDefault();if(!event.repeat)action('back');}
   else if(event.key==='Tab'){
    event.preventDefault();const buttons=Array.from(root.current?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)')??[]);
    if(!buttons.length)return;
    const index=buttons.indexOf(document.activeElement as HTMLButtonElement);
    buttons[index<0?(event.shiftKey?buttons.length-1:0):(index+(event.shiftKey?-1:1)+buttons.length)%buttons.length]?.focus();
   }else if(pressed.startsWith('Arrow')){event.preventDefault();action(pressed.slice(5).toLowerCase());}
   else if(pressed==='Enter'||pressed===' '){event.preventDefault();if(!event.repeat)action('confirm');}
  };
  window.addEventListener('keydown',key,true);return()=>{unsubscribe();window.removeEventListener('keydown',key,true);};
 },[confirm]);
 useEffect(()=>{if(session?.state==='ended'){notify(t('La partida terminó.'));requestAnimationFrame(()=>document.querySelector<HTMLElement>('.game-card.selected')?.focus({preventScroll:true}));}if(session?.state==='error')notify(t(session.message));},[session?.state]);
 if(!active||!session)return null;
 return <>
  {session.background&&!session.menu&&<button data-nav className="session-banner" onClick={()=>void run(()=>host().OpenGameMenu())} disabled={busy}><Gamepad2 size={22} aria-hidden="true"/><span><small>{t(session.state==='closing'?'Cerrando…':'En ejecución')}</small><strong dir="auto">{session.title}</strong></span><span className="session-banner-hint"><kbd>F10</kbd>{t('Menú del juego')}<ArrowRight size={18} aria-hidden="true"/></span></button>}
  {session.menu&&!session.overlay&&<SessionMenu session={session} game={games.find(game=>game.id===session.gameId)} confirm={confirm} busy={busy} error={error} menuRef={root} onResume={()=>void run(()=>host().ResumeGame())} onHome={()=>void run(()=>host().ReturnToGemu())} onAssist={()=>void run(async()=>{const open=host().OpenInputAssistant;if(!open)throw new Error(t('Reinicia Gemu para usar la asistencia de escritorio.'));await open();})} onClose={()=>{setError('');setConfirm(true);interfaceAudio.play('open');}} onCancel={()=>setConfirm(false)} onConfirm={()=>void run(async()=>{await host().CloseGame();setConfirm(false);})}/>}
 </>;
}
