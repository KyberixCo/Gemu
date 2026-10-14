import {useEffect,useRef,useState,type ReactNode} from 'react';
import {libraryHost} from './library';
import {setStorageProfile,startupAction,type GemuProfile,type ProfileHost,type SteamAccount} from './profiles';
import {moveFocus} from './navigation';
import {subscribeGamepad} from './features/feedback/gamepad';
import {navigationKey} from './keyboard-bindings';
import {useLocalization} from './localization';
import {browserController} from './features/controllers/providers';
import {controllerAssociations,controllerRegistryKey,readControllerRegistry} from './features/controllers/registry';

export function ProfileGate({children}:{children:(profile:GemuProfile,manage:()=>void)=>ReactNode}){
 const {language,setLanguage,t}=useLocalization();
 const [profiles,setProfiles]=useState<GemuProfile[]>([]),[active,setActive]=useState<GemuProfile|null>(null);
 const [suggestedPlayer,setSuggestedPlayer]=useState('');
 const [mode,setMode]=useState<'loading'|'language'|'create'|'select'|'active'>('loading'),[step,setStep]=useState(0),[name,setName]=useState(''),[error,setError]=useState(''),[busy,setBusy]=useState(false),[attempt,setAttempt]=useState(0);
 const [steamAccounts,setSteamAccounts]=useState<SteamAccount[]>([]),[selectedSteamAccount,setSelectedSteamAccount]=useState<SteamAccount>(),[steamAccountsLoading,setSteamAccountsLoading]=useState(false),[steamAccountsError,setSteamAccountsError]=useState('');
 const root=useRef<HTMLElement>(null);
 const host=()=>libraryHost() as ReturnType<typeof libraryHost>&ProfileHost;
 const enter=async(profile:GemuProfile)=>{setBusy(true);setError('');try{await host().SelectProfile(profile.id);setStorageProfile(profile.id);setActive(profile);setMode('active');}catch(e){setError(String(e));}finally{setBusy(false);}};
 useEffect(()=>{let cancelled=false;setMode('loading');setError('');
  void (async()=>{try{const items=await host().GetProfiles();if(cancelled)return;setProfiles(items);const action=startupAction(items);if(action==='enter'){await host().SelectProfile(items[0].id);if(cancelled)return;setStorageProfile(items[0].id);setActive(items[0]);setMode('active');}else setMode(items.length?'select':'language');}catch(e){if(!cancelled){setError(String(e));setMode(profiles.length?'select':'language');}}})();
  return()=>{cancelled=true;};
 },[attempt]);
 useEffect(()=>{if(mode!=='create')return;let cancelled=false;setSteamAccountsLoading(true);setSteamAccountsError('');
  void (async()=>{try{const accounts=await host().GetSteamAccounts();if(!cancelled)setSteamAccounts(accounts);}catch(reason){if(!cancelled){setSteamAccounts([]);setSteamAccountsError(String(reason));}}finally{if(!cancelled)setSteamAccountsLoading(false);}})();
  return()=>{cancelled=true;};
 },[mode]);
 useEffect(()=>{if(mode!=='active')root.current?.querySelector<HTMLElement>('input,button:not(:disabled)')?.focus();},[mode,step,busy]);
 useEffect(()=>{
  if(mode==='active')return;
  const action=(name:string)=>{if(name==='confirm'){if(document.activeElement instanceof HTMLInputElement)root.current?.querySelector<HTMLButtonElement>('button.primary')?.click();else(document.activeElement as HTMLElement)?.click();}else if(name==='back'){if(mode==='create'&&step)setStep(0);else if(profiles.length)setMode('select');}else if(root.current)moveFocus(name as 'left',root.current);};
  const key=(e:KeyboardEvent)=>{const value=navigationKey(e);if(value==='Escape'){e.preventDefault();action('back');}else if(value==='Enter'||value===' '&&!(e.target instanceof HTMLInputElement)){e.preventDefault();if(!e.repeat)action('confirm');}else if(value.startsWith('Arrow')&&!(e.target instanceof HTMLInputElement)){e.preventDefault();action(value.slice(5).toLowerCase());}};
  let knownControllers:ReturnType<typeof readControllerRegistry>=[];
  try{knownControllers=readControllerRegistry(JSON.parse(localStorage.getItem(controllerRegistryKey)||'[]'));}catch{/* No saved controller identities. */}
  const unsubscribe=subscribeGamepad({action,snapshot:pad=>setSuggestedPlayer(pad?controllerAssociations.resolve(browserController(pad.index,pad.id),knownControllers)?.playerId??'':'')});
  window.addEventListener('keydown',key);return()=>{window.removeEventListener('keydown',key);unsubscribe();};
 },[mode,step,profiles]);
 const create=async()=>{setBusy(true);setError('');try{const profile=await host().CreateProfileWithAvatar(name,selectedSteamAccount?.avatar??'');setProfiles([...profiles,profile]);setMode('select');await enter(profile);}catch(e){setError(String(e));}finally{setBusy(false);}};
 if(mode==='active'&&active)return children(active,()=>{setMode('select');setError('');});
 return <section className="profile-gate" ref={root} aria-label={t('Perfiles de Gemu')}><div className="profile-card"><span className="brand">{"g"}<span>.</span></span><p className="eyebrow">{t(profiles.length?'PERFILES DE GEMU':'CONFIGURACIÓN INICIAL')}</p>
 {mode==='select'&&profiles.some(player=>player.id===suggestedPlayer)&&<p role="status">{t('Jugador sugerido por el mando')}: {profiles.find(player=>player.id===suggestedPlayer)?.name}</p>}
 {mode==='loading'?<p role="status">{t('Cargando perfiles…')}</p>:mode==='language'?<><h1>{t('Elige el idioma de Gemu')}</h1><p>{t('La aplicación usará este idioma para la interfaz y para priorizar metadatos y arte compatibles.')}</p><div className="profile-list"><button data-nav aria-pressed={language==='es'} onClick={()=>setLanguage('es')}>{t("Español")}</button><button data-nav aria-pressed={language==='en'} onClick={()=>setLanguage('en')}>{t("English")}</button></div><button data-nav className="primary" onClick={()=>setMode('create')}>{t('Continuar')}</button></>:mode==='create'?<><p>{t('Paso')} {step+1} {t('de')} 2</p><h1>{t(step?'Tu espacio está listo':'Crea tu perfil de jugador')}</h1>{step?<><h2 className="profile-preview"><span className="profile-avatar">{selectedSteamAccount?.avatar?<img src={selectedSteamAccount.avatar} alt=""/>:name.trim().slice(0,2).toLocaleUpperCase()}</span>{name.trim()}</h2><p>{t('Tu biblioteca, favoritos, ajustes y logros estarán separados de los demás jugadores. Al importar Steam podrás elegir la cuenta que pertenece a este perfil.')}</p><button data-nav className="primary" disabled={busy} onClick={()=>void create()}>{t(busy?'Creando…':'Crear perfil y entrar')}</button><button data-nav disabled={busy} onClick={()=>setStep(0)}>{t('Editar nombre')}</button></>:<><p>{t('¿Cómo quieres aparecer en Gemu?')}</p>{steamAccountsLoading?<p role="status">{t('Buscando cuentas de Steam disponibles…')}</p>:steamAccounts.length>0?<><p>{t('Elige una cuenta para sugerir su nombre e imagen. Podrás editar el nombre antes de crear el perfil.')}</p><div className="steam-account-list">{steamAccounts.map(account=><button data-nav type="button" key={account.id} aria-pressed={selectedSteamAccount?.id===account.id} disabled={busy} onClick={()=>{setSelectedSteamAccount(account);setName(account.name);}}><span className="profile-avatar">{account.avatar?<img src={account.avatar} alt=""/>:account.name.slice(0,2).toLocaleUpperCase()}</span><strong>{account.name}</strong></button>)}</div>{steamAccountsError&&<p role="status">{t('No se pudieron leer las imágenes locales de Steam:')} {t(steamAccountsError)}</p>}</>:<p>{t(steamAccountsError?'No se pudieron detectar cuentas Steam. Puedes crear el perfil manualmente.':'No se encontraron cuentas Steam locales. Puedes crear el perfil manualmente.')}</p>}<label>{t('Nombre del jugador')}<input data-nav autoComplete="off" maxLength={32} value={name} onChange={e=>setName(e.target.value)} disabled={busy}/></label>{selectedSteamAccount&&<button data-nav type="button" onClick={()=>setSelectedSteamAccount(undefined)}>{t('Quitar sugerencia de Steam')}</button>}<button data-nav className="primary" disabled={!name.trim()||busy} onClick={()=>{setError('');setStep(1);}}>{t('Continuar')}</button>{profiles.length>0&&<button data-nav onClick={()=>setMode('select')}>{t('Volver a los perfiles')}</button>}</>}</>:<><h1>{t('¿Quién va a jugar?')}</h1><p>{t('Cada jugador tiene su biblioteca y su cuenta de Steam.')}</p><div className="profile-list">{profiles.map(p=><button data-nav key={p.id} disabled={busy} onClick={()=>void enter(p)}><span className="profile-avatar">{p.avatar?<img src={p.avatar} alt=""/>:p.name.slice(0,2).toLocaleUpperCase()}</span><strong>{p.name}</strong>{p.id===active?.id&&<span>{t('Perfil actual')}</span>}</button>)}</div><button data-nav className="primary" disabled={busy} onClick={()=>{setName('');setSelectedSteamAccount(undefined);setStep(0);setError('');setMode('create');}}>{t('Crear otro perfil')}</button></>}
 {error&&<div role="alert"><p>{t(error)}</p>{!profiles.length&&(mode==='select'||mode==='language')&&<button data-nav onClick={()=>setAttempt(n=>n+1)}>{t('Reintentar conexión')}</button>}</div>}
 </div></section>;
}
