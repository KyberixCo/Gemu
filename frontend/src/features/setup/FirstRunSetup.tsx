import {ActionHint} from '../../ActionHint';
import {useEffect,useRef,useState,type ReactNode} from 'react';
import {libraryHost,type SourceInfo,type Snapshot} from '../../library';
import {profileKey} from '../../profiles';
import {loadManagementConfig,saveManagementConfig} from '../management/management-config';
import {ProviderSetupWizard} from '../management/ProviderSetupWizard';
import {SourceWizard} from '../management/SourceWizard';
import {WizardFrame} from '../management/WizardFrame';
import {useTVBack} from '../management/TVControls';
import {useWizardNavigation} from './useWizardNavigation';
import {setupComplete,shouldOfferSetup} from './policy';
import type {ContentProviderId} from '../management/content-providers';
import {ArtworkFlow} from '../management/ArtworkFlow';
import {useLocalization} from '../../localization';

export function FirstRunSetup({children}:{children:ReactNode}){
 const {t}=useLocalization();
 const [done,setDone]=useState(()=>setupComplete()),[loading,setLoading]=useState(!done);
 const [config,setConfig]=useState(loadManagementConfig),[sources,setSources]=useState<SourceInfo[]>([]),[library,setLibrary]=useState<Snapshot>({version:1,sources:[],games:[]});
 const [step,setStep]=useState(0),[provider,setProvider]=useState<'screenscraper'|'steamgriddb'>(),[source,setSource]=useState<SourceInfo>(),[values,setValues]=useState<Record<string,string>>({}),[error,setError]=useState(''),[improveArtwork,setImproveArtwork]=useState(false);
 const [sourceBusy,setSourceBusy]=useState(false);
 const root=useRef<HTMLElement>(null);
 const loadSequence=useRef(0),forced=useRef(false);
 useWizardNavigation(root,!done);
 useTVBack(()=>{if(improveArtwork)return;if(source){if(!sourceBusy)setSource(undefined);return;}if(provider){setProvider(undefined);setValues({});return;}if(step)setStep(step-1);},!done);
 const load=async()=>{
  const id=++loadSequence.current;
  setLoading(true);setError('');
  try{const host=libraryHost(),[adapters,snapshot]=await Promise.all([host.GetSources(),host.GetLibrary()]);if(id!==loadSequence.current)return;setSources(adapters);setLibrary(snapshot);if(!forced.current&&!shouldOfferSetup(setupComplete(),snapshot))setDone(true);}
  catch(reason){if(id===loadSequence.current)setError(String(reason));}finally{if(id===loadSequence.current)setLoading(false);}
 };
 useEffect(()=>{if(!done)void load();const open=()=>{forced.current=true;setStep(0);setProvider(undefined);setSource(undefined);setValues({});setConfig(loadManagementConfig());setDone(false);void load();};window.addEventListener('gemu-open-setup',open);return()=>{loadSequence.current++;window.removeEventListener('gemu-open-setup',open);};},[]);
 const finish=()=>{
  try{saveManagementConfig(config);localStorage.setItem(profileKey('gemu.setup.v1'),'complete');setDone(true);}
  catch{setError('No se pudo guardar la configuración. Puedes reintentar o entrar sin conservar las preferencias.');}
 };
 if(done)return <>{children}</>;
 const steps=['Tu consola','Arte y metadatos','Tus juegos','Revisión'];
 return <section className="setup-gate" ref={root} aria-label={t('Primera configuración de Gemu')}><div className="setup-panel management-tv">
 <header className="tv-header"><span className="brand">{"g"}<span>.</span></span><p className="eyebrow">{t('CONFIGURACIÓN INICIAL')}</p></header>
 <div className="tv-content">{loading?<p role="status">{t('Preparando tu configuración…')}</p>:improveArtwork?<ArtworkFlow bulk entries={library.games.filter(entry=>!entry.metadata.cover||!entry.metadata.background||!entry.metadata.logo||!entry.metadata.icon)} settings={config.contentProviders} credentials={{}} onSaved={snapshot=>{setLibrary(snapshot);setImproveArtwork(false);}} onClose={()=>setImproveArtwork(false)} onConfigure={()=>{setImproveArtwork(false);setStep(1);}}/>:provider?<ProviderSetupWizard key={provider} provider={provider} preferences initialEnabled={config.contentProviders[provider].enabled} initialPriority={config.contentProviders[provider].priority} values={values} onChange={setValues} onCancel={()=>{setProvider(undefined);setValues({});}} onComplete={result=>{setConfig(current=>({...current,contentProviders:{...current.contentProviders,[provider]:{enabled:result.enabled,priority:result.priority}}}));setProvider(undefined);setValues({});}}/>:source?<SourceWizard key={source.id} sourceId={source.id} name={source.name} enabled={library.sources.some(item=>item.id===source.id)} path={library.sources.find(item=>item.id===source.id)?.root??''} contentSettings={config.contentProviders} credentials={{}} onBusyChange={setSourceBusy} onComplete={snapshot=>{setLibrary(snapshot);setSource(undefined);setSourceBusy(false);setStep(3);}}/>:
 <WizardFrame title={t('Preparar Gemu')} steps={steps.map(text=>t(text))} current={step} actions={<><button data-nav className="manager-secondary" onClick={()=>step?setStep(step-1):finish()}>{t(step?'Atrás':'Configurar más tarde')}</button>{step<3?<button data-nav className="primary" onClick={()=>setStep(step+1)}>{t('Continuar')}</button>:<button data-nav className="primary" onClick={finish}>{t('Guardar y entrar en Gemu')}</button>}</>}>
 {step===0&&<><p>{t('Tu perfil ya está creado. Ahora puedes conectar el arte y traer tus juegos.')}</p><p>{t('Cada servicio tiene su propio asistente. Puedes dejar cualquier conexión pendiente y configurarla más tarde desde Gestión.')}</p></>}
 {step===1&&<><p>{t('Steam aporta el arte básico. Configura SteamGridDB para ampliarlo o ScreenScraper para obtener metadatos y arte.')}</p><div className="tv-setting-list">{(['steamgriddb','screenscraper'] as const).map(id=><button data-nav className="tv-row" key={id} onClick={()=>{setValues({});setProvider(id);}}><strong>{id==='steamgriddb'?"SteamGridDB":"ScreenScraper"}</strong><span>{t(config.contentProviders[id].enabled?'Configurado · revisar':'Abrir asistente')}</span></button>)}</div><p>{t('El acceso de aplicación a ScreenScraper lo proporciona Gemu. Solo necesitas configurar tu cuenta personal.')}</p></>}
 {step===2&&<><p>{t('El asistente de biblioteca detecta instalaciones, configura la cuenta y prepara los juegos antes de publicarlos.')}</p><div className="tv-setting-list">{sources.map(adapter=><button data-nav className="tv-row" key={adapter.id} onClick={()=>setSource(adapter)}><strong>{adapter.name}</strong><span>{t('Configurar biblioteca')}</span></button>)}</div>{!sources.length&&<p>{t('No hay fuentes disponibles. Reinicia el host actualizado o configura la biblioteca más tarde.')}</p>}</>}
 {step===3&&<><p>{library.games.length} {t('juegos en tu biblioteca.')}</p><p>{(Object.keys(config.contentProviders) as ContentProviderId[]).filter(id=>config.contentProviders[id].enabled).map(id=>id==='steamgriddb'?'SteamGridDB':id==='screenscraper'?'ScreenScraper':id==='manual'?'Archivos propios':id).join(' · ')||t('Sin proveedores adicionales activos')}</p><p>{t('Las conexiones pendientes se pueden retomar desde Gestión. Puedes volver a ejecutar este asistente cuando lo necesites.')}</p>{library.games.some(entry=>!entry.metadata.cover||!entry.metadata.background||!entry.metadata.logo||!entry.metadata.icon)&&<><p>{t('Si alguna imagen quedó pendiente, puedes buscarla ahora y revisarla antes de guardarla.')}</p><button data-nav className="manager-secondary" disabled={!Object.values(config.contentProviders).some(value=>value.enabled)} onClick={()=>setImproveArtwork(true)}>{t('Completar arte faltante')}</button>{!Object.values(config.contentProviders).some(value=>value.enabled)&&<button data-nav className="manager-secondary" onClick={()=>setStep(1)}>{t('Configurar proveedores')}</button>}</>}</>}
 {error&&<><p role="status">{t(error)}</p><button data-nav className="manager-secondary" onClick={()=>void load()}>{t("Reintentar conexión")}</button><button data-nav className="manager-secondary" onClick={()=>setDone(true)}>{t("Entrar sin guardar")}</button></>}
 </WizardFrame>}</div>
 <footer className="tv-footer"><span><ActionHint action="confirm"/> {t('Seleccionar')} · <ActionHint action="back"/> {t('Atrás')}</span>{(source||provider)&&<button data-nav className="manager-secondary" disabled={sourceBusy} onClick={()=>{setSource(undefined);setProvider(undefined);setValues({});}}>{t('Volver al asistente principal')}</button>}</footer>
 </div></section>;
}
