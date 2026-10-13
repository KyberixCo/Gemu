import {useEffect, useRef, useState} from 'react';
import {TVRow, useTVBack} from './TVControls';
import {contentHost,libraryHost,type Preview,type Draft,type Snapshot} from '../../library';
import type {ContentSettings} from './ContentProviders';
import {contentRequests} from './content-requests';
import {CredentialEditor} from './CredentialEditor';

type Step='idle'|'manual'|'discovering'|'profile'|'enriching'|'review'|'saving'|'done'|'error';
export function SourceWizard({sourceId,name,enabled,path,onComplete,contentSettings,credentials,onBusyChange}:{sourceId:string;name:string;enabled:boolean;path:string;onComplete:(snapshot:Snapshot)=>void;contentSettings:ContentSettings;credentials:Record<string,string>;onBusyChange?:(busy:boolean)=>void}) {
 const [step,setStep]=useState<Step>('idle'),[preview,setPreview]=useState<Preview>(),[draft,setDraft]=useState<Draft>(),[error,setError]=useState('');
 const [manualPath,setManualPath]=useState(path),[scanPath,setScanPath]=useState(''),[profileId,setProfileId]=useState(''),[apiKey,setApiKey]=useState('');
 const request=useRef(0),root=useRef<HTMLDivElement>(null),pendingToken=useRef('');
 const discard=()=>{if(pendingToken.current){void libraryHost().DiscardLibraryImport(pendingToken.current).catch(()=>{});pendingToken.current='';}};
 const cancel=()=>{request.current++;discard();setStep('idle');setPreview(undefined);setDraft(undefined);setError('');setApiKey('');};
 useTVBack(()=>{if(step!=='saving'&&!document.querySelector('.tv-editor'))cancel();},step!=='idle',true);
 useEffect(()=>{onBusyChange?.(step==='saving');},[step,onBusyChange]);
 useEffect(()=>()=>{request.current++;discard();},[]);
 useEffect(()=>{root.current?.querySelector<HTMLElement>('[data-nav]:not(:disabled)')?.focus({preventScroll:true});},[step]);
 const openManual=()=>{request.current++;discard();setError('');setManualPath(preview?.root||scanPath||path);setStep('manual');};
 const discover=async(candidate='')=>{
  const id=++request.current;discard();setScanPath(candidate);setStep('discovering');setError('');setDraft(undefined);
  try{const next=await libraryHost().PreviewSource(sourceId,candidate);if(request.current===id){setPreview(next);setProfileId(next.profiles.length===1?next.profiles[0].id:'');setStep('profile');}}
  catch(reason){if(request.current===id){setError(String(reason));setStep('error');}}
 };
 const enrich=async()=>{
  if(!preview)return;const id=++request.current;setStep('enriching');setError('');
  try{
   const providers=contentRequests(contentSettings,credentials);
   const host=providers.length?contentHost():libraryHost();
   const next=await host.PrepareLibraryImport({sourceId,root:preview.root,profileId,apiKey,contentProviders:providers});
   if(request.current!==id){void libraryHost().DiscardLibraryImport(next.token);return;}
   pendingToken.current=next.token;setDraft(next);setApiKey('');setStep('review');
  }catch(reason){if(request.current===id){setError(String(reason));setStep('profile');}}
 };
 const finish=async()=>{
  if(!draft)return;setStep('saving');setError('');
  try{const snapshot=await libraryHost().CommitLibraryImport(draft.token);pendingToken.current='';onComplete(snapshot);setStep('done');}
  catch(reason){setError(String(reason));setStep('review');}
 };
 return <div ref={root} className="source-wizard">
 {step==='idle'?<>
  <p>{enabled?`${name} está configurado. Repite el asistente para actualizar sus juegos, metadatos y perfil.`:`Configura ${name} para incorporar tus juegos a la biblioteca.`}</p>
  <TVRow label={enabled?`Actualizar ${name}`:`Configurar ${name}`} value="Detección automática · recomendada" onClick={()=>void discover()}/>
  <TVRow label="Configurar carpeta manualmente" value="Indicar otra ubicación" onClick={openManual}/>
 </>:<>
  <ol className="source-wizard-steps" aria-label="Pasos de importación"><li aria-current={['manual','discovering','error'].includes(step)?'step':undefined}>1. Juegos</li><li aria-current={step==='profile'?'step':undefined}>2. Cuenta</li><li aria-current={['enriching','review'].includes(step)?'step':undefined}>3. Metadatos y logros</li><li aria-current={['saving','done'].includes(step)?'step':undefined}>4. Biblioteca</li></ol>
  {step==='manual'&&<form onSubmit={e=>{e.preventDefault();if(!manualPath.trim()){setError('Indica una carpeta para continuar.');return;}void discover(manualPath.trim());}}>
   <label className="manager-field"><span>Carpeta de {name}</span><input data-nav value={manualPath} onChange={e=>{setManualPath(e.target.value);setError('');}} autoComplete="off"/></label>
   <p>Introduce la ruta completa de la instalación o biblioteca.</p><button data-nav className="primary">Validar carpeta</button>
  </form>}
  <div role="status" aria-live="polite">
   {step==='discovering'&&<p>Obteniendo el listado de juegos y los perfiles locales…</p>}
   {step==='enriching'&&<p>Obteniendo metadatos, guardando arte y consultando logros…</p>}
   {step==='saving'&&<p>Guardando tu biblioteca…</p>}
   {step==='error'&&<h3>No pudimos consultar esta fuente</h3>}
   {error&&<p>{error}</p>}
  </div>
  {step==='profile'&&preview&&<>
   <h3>Asociar cuenta de {name}</h3>
   <p>{preview.games.filter(g=>g.Installations.some(i=>i.Installed)).length} juegos instalados encontrados.</p>
   <div className="tv-options" aria-label={`Cuenta de ${name}`}><button data-nav aria-pressed={profileId===''} onClick={()=>setProfileId('')}>Sin cuenta asociada</button>{preview.profiles.map(p=><button data-nav key={p.id} aria-pressed={profileId===p.id} onClick={()=>setProfileId(p.id)}>{p.name}</button>)}</div>
   <CredentialEditor provider="steam" fields={[{name:'apiKey',label:'Clave API de Steam · opcional'}]} values={{apiKey}} onChange={values=>setApiKey(values.apiKey??'')}/>
   <p>Con una clave se consultan el perfil y los logros por API. Si no están disponibles, se conserva el perfil local y los logros quedan pendientes.</p>
   <p className="manager-hint">{contentRequests(contentSettings,credentials).length?'El arte se ampliará con '+contentRequests(contentSettings,credentials).map(request=>request.id==='steamgriddb'?'SteamGridDB':'ScreenScraper').join(' → ')+'. Se usan los valores nuevos o las credenciales guardadas del perfil.':'Puedes ampliar el arte habilitando SteamGridDB o ScreenScraper en Información y arte.'}</p>
   <button data-nav className="primary" onClick={()=>void enrich()}>Obtener metadatos y continuar</button>
  </>}
  {step==='review'&&draft&&<>
   <h3>{draft.games.length} juegos preparados para tu biblioteca</h3>
   <ul className="source-results" data-nav tabIndex={0} aria-label="Juegos preparados">{draft.games.map(entry=><li key={entry.game.ID}><strong>{entry.game.Title}</strong><span>{entry.metadata.status==='ready'?'Metadatos listos':entry.metadata.status==='cached'?'Metadatos guardados':'Metadatos pendientes'} · {entry.player.status==='ready'?'Logros listos':entry.player.status==='cached'?'Logros guardados':'Logros pendientes'} · {entry.metadata.artworks?.filter(art=>art.kind==='background').length??0} fondos externos</span></li>)}</ul>
   {!!draft.diagnostics.length&&<details className="source-wizard-diagnostics"><summary data-nav>Incidencias ({draft.diagnostics.length})</summary>{draft.diagnostics.map((message,i)=><p key={i}>{message}</p>)}</details>}
   <p>Finalizar guardará los juegos y los mostrará en Jugar. Puedes importar aunque algún dato esté pendiente.</p>
  </>}
  {step==='done'&&<><h3>Tu biblioteca está lista</h3><p>Los juegos, el arte y el perfil asociado están guardados. Ya puedes verlos en Jugar.</p></>}
  <div className="tv-subnav">
   {step==='review'&&<button data-nav className="primary" onClick={()=>void finish()}>Finalizar e incorporar juegos</button>}
   {step==='error'&&<button data-nav onClick={()=>void discover(scanPath)}>Reintentar</button>}
   {(step==='error'||step==='profile'||step==='review')&&<button data-nav onClick={openManual}>Cambiar carpeta</button>}
   <button data-nav disabled={step==='saving'} onClick={cancel}>{step==='done'?'Cerrar asistente':'Cancelar'}</button>
  </div>
 </>}
 </div>;
}
