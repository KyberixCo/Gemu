import {useEffect,useRef,useState,type ReactNode} from 'react';
import {libraryHost} from './library';
import {setStorageProfile,startupAction,type GemuProfile,type ProfileHost} from './profiles';
import {moveFocus} from './navigation';
import {GamepadRepeater} from './features/feedback/input';

export function ProfileGate({children}:{children:(profile:GemuProfile,manage:()=>void)=>ReactNode}){
 const [profiles,setProfiles]=useState<GemuProfile[]>([]),[active,setActive]=useState<GemuProfile|null>(null);
 const [mode,setMode]=useState<'loading'|'create'|'select'|'active'>('loading'),[step,setStep]=useState(0),[name,setName]=useState(''),[error,setError]=useState(''),[busy,setBusy]=useState(false),[attempt,setAttempt]=useState(0);
 const root=useRef<HTMLElement>(null);
 const host=()=>libraryHost() as ReturnType<typeof libraryHost>&ProfileHost;
 const enter=async(profile:GemuProfile)=>{setBusy(true);setError('');try{await host().SelectProfile(profile.id);setStorageProfile(profile.id);setActive(profile);setMode('active');}catch(e){setError(String(e));}finally{setBusy(false);}};
 useEffect(()=>{let cancelled=false;setMode('loading');setError('');
  void (async()=>{try{const items=await host().GetProfiles();if(cancelled)return;setProfiles(items);const action=startupAction(items);if(action==='enter'){await host().SelectProfile(items[0].id);if(cancelled)return;setStorageProfile(items[0].id);setActive(items[0]);setMode('active');}else setMode(action);}catch(e){if(!cancelled){setError(String(e));setMode('select');}}})();
  return()=>{cancelled=true;};
 },[attempt]);
 useEffect(()=>{if(mode!=='active')root.current?.querySelector<HTMLElement>('input,button:not(:disabled)')?.focus();},[mode,step,busy]);
 useEffect(()=>{
  if(mode==='active')return;
  const repeater=new GamepadRepeater();let frame=0;
  const action=(name:string)=>{if(name==='confirm'){if(document.activeElement instanceof HTMLInputElement)root.current?.querySelector<HTMLButtonElement>('button.primary')?.click();else(document.activeElement as HTMLElement)?.click();}else if(name==='back'){if(mode==='create'&&step)setStep(0);else if(profiles.length)setMode('select');}else if(root.current)moveFocus(name as 'left',root.current);};
  const key=(e:KeyboardEvent)=>{if(e.key==='Escape'){e.preventDefault();action('back');}else if(e.key==='Enter'&&e.target instanceof HTMLInputElement){e.preventDefault();action('confirm');}else if(e.key.startsWith('Arrow')&&!(e.target instanceof HTMLInputElement)){e.preventDefault();action(e.key.slice(5).toLowerCase());}};
  const poll=(now:number)=>{const pad=navigator.getGamepads?.().find(p=>p?.connected);if(pad&&document.hasFocus()){const next=pad.buttons[0]?.pressed?'confirm':pad.buttons[1]?.pressed?'back':pad.buttons[12]?.pressed||pad.axes[1]<-.5?'up':pad.buttons[13]?.pressed||pad.axes[1]>.5?'down':pad.buttons[14]?.pressed||pad.axes[0]<-.5?'left':pad.buttons[15]?.pressed||pad.axes[0]>.5?'right':'';if(repeater.accept(next,now))action(next);}else repeater.reset();frame=requestAnimationFrame(poll);};
  window.addEventListener('keydown',key);frame=requestAnimationFrame(poll);return()=>{window.removeEventListener('keydown',key);cancelAnimationFrame(frame);};
 },[mode,step,profiles]);
 const create=async()=>{setBusy(true);setError('');try{const profile=await host().CreateProfile(name);setProfiles([...profiles,profile]);setMode('select');await enter(profile);}catch(e){setError(String(e));}finally{setBusy(false);}};
 if(mode==='active'&&active)return children(active,()=>{setMode('select');setError('');});
 return <section className="profile-gate" ref={root} aria-label="Perfiles de Gemu"><div className="profile-card"><span className="brand">g<span>.</span></span><p className="eyebrow">{profiles.length?'PERFILES DE GEMU':'CONFIGURACIÓN INICIAL'}</p>
 {mode==='loading'?<p role="status">Cargando perfiles…</p>:mode==='create'?<><p>Paso {step+1} de 2</p><h1>{step?'Tu espacio está listo':'Crea tu perfil de jugador'}</h1>{step?<><h2>{name.trim()}</h2><p>Tu biblioteca, favoritos, ajustes y logros estarán separados de los demás jugadores. Al importar Steam podrás elegir la cuenta que pertenece a este perfil.</p><button data-nav className="primary" disabled={busy} onClick={()=>void create()}>{busy?'Creando…':'Crear perfil y entrar'}</button><button data-nav disabled={busy} onClick={()=>setStep(0)}>Editar nombre</button></>:<><p>¿Cómo quieres aparecer en Gemu?</p><label>Nombre del jugador<input data-nav autoComplete="off" maxLength={32} value={name} onChange={e=>setName(e.target.value)} disabled={busy}/></label><button data-nav className="primary" disabled={!name.trim()||busy} onClick={()=>{setError('');setStep(1);}}>Continuar</button>{profiles.length>0&&<button data-nav onClick={()=>setMode('select')}>Volver a los perfiles</button>}</>}</>:<><h1>¿Quién va a jugar?</h1><p>Cada jugador tiene su biblioteca y su cuenta de Steam.</p><div className="profile-list">{profiles.map(p=><button data-nav key={p.id} disabled={busy} onClick={()=>void enter(p)}><span className="avatar">{p.name.slice(0,2).toLocaleUpperCase()}</span><strong>{p.name}</strong>{p.id===active?.id&&<span>Perfil actual</span>}</button>)}</div><button data-nav className="primary" disabled={busy} onClick={()=>{setName('');setStep(0);setError('');setMode('create');}}>Crear otro perfil</button></>}
 {error&&<div role="alert"><p>{error}</p>{!profiles.length&&mode==='select'&&<button data-nav onClick={()=>setAttempt(n=>n+1)}>Reintentar conexión</button>}</div>}
 </div></section>;
}
