import {useEffect,useRef,useState} from 'react';
import type {Snapshot} from '../../library';
import {useLocalization} from '../../localization';
import {TVField,useTVBack} from './TVControls';
import {WizardFrame} from './WizardFrame';
import {chooseCollections,installSize,isInstalling,retroHost,type RetroConfig,type RetroJob,type RetroPlan,type RetroPreview,type RetroStatus} from './retroarch';

const blank:RetroConfig={executable:'',systemDirectory:'',collections:[],fullscreen:true,rewind:false,language:'es',joypadDriver:'auto'};
const steps=['Detectar RetroArch','Elegir consolas','Instalar componentes','Carpetas de juegos','Experiencia y controles','Revisar y guardar'];
const stages:Record<string,string>={starting:'Preparando instalación…',downloading:'Descargando…',extracting:'Extrayendo y comprobando…',publishing:'Finalizando instalación…',ready:'Instalación lista',cancelled:'Instalación cancelada',error:'La instalación no terminó'};
export function RetroArchWizard({profileID,initialExecutable='',onSaved,onCancel,onBusyChange}:{profileID:string;initialExecutable?:string;onSaved:(snapshot:Snapshot)=>void;onCancel:()=>void;onBusyChange?:(busy:boolean)=>void}){
 const {t,language}=useLocalization();
 const [step,setStep]=useState(0),[status,setStatus]=useState<RetroStatus>(),[config,setConfig]=useState<RetroConfig>(blank),[systems,setSystems]=useState<string[]>([]),[installApp,setInstallApp]=useState(false);
 const [manual,setManual]=useState(initialExecutable),[plan,setPlan]=useState<RetroPlan>(),[job,setJob]=useState<RetroJob>(),[preview,setPreview]=useState<RetroPreview>(),[error,setError]=useState(''),[busy,setBusy]=useState(false),[saved,setSaved]=useState(false),[saving,setSaving]=useState(false),[previewPage,setPreviewPage]=useState(0);
 const live=useRef(true),sequence=useRef(0),planRef=useRef(''),previewRef=useRef(''),jobRef=useRef(''),pollFailures=useRef(0),poll=useRef<ReturnType<typeof setTimeout>|undefined>(undefined),root=useRef<HTMLDivElement>(null);
 const installing=isInstalling(job),locked=saving||job?.state==='publishing';
 const fail=(reason:unknown)=>{if(live.current)setError(String(reason instanceof Error?reason.message:reason));};
 const discardPreview=()=>{if(previewRef.current){void retroHost().DiscardRetroArchPreview(profileID,previewRef.current).catch(()=>{});previewRef.current='';}setPreview(undefined);};
 const discardPlan=()=>{if(planRef.current){void retroHost().DiscardRetroArchInstallPlan(profileID,planRef.current).catch(()=>{});planRef.current='';}setPlan(undefined);};
 const detect=async(path=manual)=>{
  const generation=++sequence.current;setBusy(true);setError('');
  try{const result=await retroHost().GetRetroArchStatus(profileID,path);if(!live.current||generation!==sequence.current)return;
   setStatus(result);const executable=result.config.executable||result.installations[0]?.executable||'';
   setConfig({...result.config,executable,joypadDriver:result.config.joypadDriver||'auto',language});setSystems(result.config.collections.map(item=>item.systemId));setInstallApp(!executable);
  }catch(reason){if(generation===sequence.current)fail(reason);}finally{if(live.current&&generation===sequence.current)setBusy(false);}
 };
 useEffect(()=>{live.current=true;void detect(initialExecutable);return()=>{live.current=false;sequence.current++;clearTimeout(poll.current);try{const host=retroHost();if(jobRef.current)void host.CancelRetroArchInstall(profileID,jobRef.current).catch(()=>{});if(previewRef.current)void host.DiscardRetroArchPreview(profileID,previewRef.current).catch(()=>{});if(planRef.current)void host.DiscardRetroArchInstallPlan(profileID,planRef.current).catch(()=>{});}catch{}};},[profileID]);
 useEffect(()=>{onBusyChange?.(Boolean(locked));return()=>onBusyChange?.(false);},[locked,onBusyChange]);
 useEffect(()=>{root.current?.querySelector<HTMLElement>('[data-nav]:not(:disabled)')?.focus({preventScroll:true});},[step,saved]);
 const back=()=>{
  if(locked)return;
  if(installing){void retroHost().CancelRetroArchInstall(profileID,job!.id).catch(fail);return;}
  sequence.current++;setBusy(false);setError('');discardPreview();discardPlan();
  if(step)setStep(step-1);else onCancel();
 };
 useTVBack(()=>{if(!document.querySelector('.tv-editor'))back();},true,true);
 const pick=async(kind:string,receive:(path:string)=>void)=>{const generation=sequence.current;try{const path=await retroHost().PickRetroArchPath(profileID,kind);if(path&&live.current&&generation===sequence.current)receive(path);}catch(reason){if(generation===sequence.current)fail(reason);}};
 const preparePlan=async()=>{
  discardPlan();setJob(undefined);jobRef.current='';const generation=++sequence.current;setStep(2);setBusy(true);setError('');
  try{const result=await retroHost().PlanRetroArchInstall(profileID,{executable:config.executable,coreDirectory:config.coreDirectory,installApp,systems});
   if(!live.current||generation!==sequence.current){void retroHost().DiscardRetroArchInstallPlan(profileID,result.token).catch(()=>{});return;}
   planRef.current=result.token;setPlan(result);
  }catch(reason){if(generation===sequence.current)fail(reason);}finally{if(live.current&&generation===sequence.current)setBusy(false);}
 };
 const followJob=async(id:string)=>{
  try{const next=await retroHost().GetRetroArchInstall(profileID,id);if(!live.current)return;pollFailures.current=0;setJob(next);
   if(next.state==='ready'){jobRef.current='';setConfig(current=>({...current,executable:next.executable}));setStep(3);setError('');}
   else if(next.state==='error'||next.state==='cancelled'){jobRef.current='';setError(next.error||t('Instalación cancelada'));discardPlan();}
   else poll.current=setTimeout(()=>void followJob(id),500);
  }catch(reason){if(live.current){fail(reason);if(++pollFailures.current<3)poll.current=setTimeout(()=>void followJob(id),2000);else{void retroHost().CancelRetroArchInstall(profileID,id).catch(()=>{});jobRef.current='';setJob(current=>current?{...current,state:'error'}:current);discardPlan();}}}
 };
 const install=async()=>{
  if(!plan||busy||installing)return;const generation=++sequence.current;setBusy(true);setError('');
  try{const next=await retroHost().StartRetroArchInstall(profileID,plan.token);planRef.current='';
   if(!live.current||generation!==sequence.current){void retroHost().CancelRetroArchInstall(profileID,next.id).catch(()=>{});return;}
   jobRef.current=next.id;setJob(next);void followJob(next.id);
  }catch(reason){if(generation===sequence.current){fail(reason);discardPlan();}}finally{if(live.current&&generation===sequence.current)setBusy(false);}
 };
 const useFolders=async()=>{const generation=++sequence.current;setBusy(true);setError('');try{const paths=await retroHost().CreateRetroArchROMFolders(profileID,systems);if(live.current&&generation===sequence.current)setConfig(current=>({...current,collections:chooseCollections(current.collections,systems).map(item=>({...item,directory:paths[item.systemId]}))}));}catch(reason){if(generation===sequence.current)fail(reason);}finally{if(live.current&&generation===sequence.current)setBusy(false);}};
 const review=async()=>{
  discardPreview();setStep(5);setBusy(true);setError('');setPreviewPage(0);const generation=++sequence.current;
  try{const result=await retroHost().PreviewRetroArch(profileID,{...config,language});
   if(!live.current||generation!==sequence.current){void retroHost().DiscardRetroArchPreview(profileID,result.token).catch(()=>{});return;}
   previewRef.current=result.token;setPreview(result);
  }catch(reason){if(generation===sequence.current)fail(reason);}finally{if(live.current&&generation===sequence.current)setBusy(false);}
 };
 const finish=async()=>{
  if(!preview?.canPublish||busy)return;setBusy(true);setSaving(true);setError('');
  try{const snapshot=await retroHost().CommitRetroArch(profileID,preview.token);previewRef.current='';if(live.current){onSaved(snapshot);setSaved(true);}}
  catch(reason){fail(reason);discardPreview();}finally{if(live.current){setBusy(false);setSaving(false);}}
 };
 const patchCollection=(id:string,patch:Partial<RetroConfig['collections'][number]>)=>setConfig(current=>({...current,collections:chooseCollections(current.collections,systems).map(item=>item.systemId===id?{...item,...patch}:item)}));
 const selectedSystems=status?.systems.filter(item=>systems.includes(item.id))??[];
 const totalSize=plan?installSize(plan):null;
 return <div className="retroarch-wizard" ref={root} aria-busy={busy||installing}>
 {saved?<><h2>{t('RetroArch está configurado')}</h2><p>{t('Tus juegos ya aparecen en la biblioteca. Puedes repetir este asistente para actualizarla.')}</p><button data-nav className="primary" onClick={onCancel}>{t('Volver')}</button></>:<WizardFrame title="RetroArch" steps={steps.map(text=>t(text))} current={step} actions={<>
  <button data-nav disabled={Boolean(locked)} onClick={back}>{t(installing?'Cancelar instalación':step?'Atrás':'Volver')}</button>
  {step===0&&status?.supported&&<button data-nav className="primary" disabled={busy||(!installApp&&!config.executable)} onClick={()=>setStep(1)}>{t(installApp?'Instalar y configurar RetroArch':'Usar esta instalación')}</button>}
  {step===1&&<button data-nav className="primary" disabled={!systems.length||busy} onClick={()=>{setConfig(current=>({...current,collections:chooseCollections(current.collections,systems)}));void preparePlan();}}>{t('Continuar')}</button>}
  {step===2&&plan&&!installing&&!busy&&<button data-nav className="primary" onClick={()=>void install()}>{t(plan.downloads.length?'Descargar e instalar':'Continuar')}</button>}
  {step===3&&<button data-nav className="primary" disabled={busy||config.collections.some(item=>!item.directory.trim())} onClick={()=>setStep(4)}>{t('Continuar')}</button>}
  {step===4&&<button data-nav className="primary" disabled={busy} onClick={()=>void review()}>{t('Revisar juegos')}</button>}
  {step===5&&preview?.canPublish&&<button data-nav className="primary" disabled={busy} onClick={()=>void finish()}>{t('Guardar e incorporar juegos')}</button>}
 </>}>
 {error&&<p role="alert">{t(error)}</p>}
 {step===0&&<>
  {busy&&<p role="status">{t('Buscando instalaciones de RetroArch…')}</p>}
  {status&&!status.supported&&<p>{t('La instalación asistida de RetroArch está disponible en Windows de 64 bits.')}</p>}
  {status?.supported&&<>
   <p>{t(status.installations.length?'Encontramos RetroArch. Elige qué instalación usar.':'No encontramos RetroArch. Gemu puede descargar y preparar la versión oficial para ti.')}</p>
   <div className="tv-setting-list">{status.installations.map(item=><button data-nav className="tv-row" key={item.executable} aria-pressed={!installApp&&config.executable===item.executable} onClick={()=>{setConfig(current=>({...current,executable:item.executable}));setInstallApp(false);}}><strong>RetroArch</strong><span>{item.executable}</span></button>)}</div>
   <button data-nav aria-pressed={installApp} onClick={()=>setInstallApp(true)}>{t('Instalar una copia administrada por Gemu')}</button>
   <TVField label={t('Carpeta de núcleos existente (opcional)')} value={config.coreDirectory??''} onChange={coreDirectory=>setConfig(current=>({...current,coreDirectory}))}/><button data-nav disabled={busy} onClick={()=>void pick('cores',coreDirectory=>setConfig(current=>({...current,coreDirectory})))}>{t('Elegir carpeta de núcleos')}</button>
   <TVField label={t('Ubicación de RetroArch')} value={manual} onChange={setManual}/>
   <div className="tv-subnav"><button data-nav disabled={busy} onClick={()=>void detect()}>{t('Buscar de nuevo')}</button><button data-nav disabled={busy} onClick={()=>void pick('executable',path=>{setManual(path);void detect(path);})}>{t('Elegir ejecutable')}</button></div>
  </>}
  {status?.diagnostics.map((message,index)=><p key={index}>{t(message)}</p>)}
  {!status&&!busy&&<button data-nav onClick={()=>void detect()}>{t('Reintentar')}</button>}
 </>}
 {step===1&&<><p>{t('Elige las consolas que vas a utilizar. Instalaremos solo los núcleos que falten.')}</p><div className="tv-choice-grid">{status?.systems.map(item=><button data-nav key={item.id} aria-pressed={systems.includes(item.id)} className="tv-module" onClick={()=>setSystems(current=>current.includes(item.id)?current.filter(id=>id!==item.id):[...current,item.id])}><strong>{item.name}</strong><span>{item.coreName}</span></button>)}</div><p>{t('Cada núcleo conserva su propia licencia. Los juegos y las BIOS los aportas tú.')}</p></>}
 {step===2&&<>
  {busy&&!job&&<p role="status">{t('Consultando los componentes oficiales…')}</p>}
  {plan&&!installing&&<><h3>{plan.version?`RetroArch ${plan.version}`:t('Preparar núcleos')}</h3><p>{t('RetroArch estable y núcleos del actualizador oficial de Libretro. No se descargarán juegos ni BIOS.')}</p><p>{t('Ubicación')}: {plan.destination}</p><p>{totalSize===null?t('Tamaño de descarga por confirmar'):t('Descarga aproximada')+': '+(totalSize/1024/1024).toFixed(1)+' MB'}</p><p>{t('Necesitas espacio adicional para descomprimir los componentes.')}</p>{plan.recommendedBytes!==undefined&&<p>{t('Espacio recomendado')}: {(plan.recommendedBytes/1024**3).toFixed(2)} GB</p>}{plan.freeBytes!==undefined&&plan.freeBytes>=0&&<p>{t('Espacio disponible')}: {(plan.freeBytes/1024**3).toFixed(2)} GB</p>}<ul>{plan.downloads.map(item=><li key={item.name}>{item.name}</li>)}</ul>{!plan.downloads.length&&<p>{t('Los componentes necesarios ya están instalados.')}</p>}</>}
  {job&&<div role="status" aria-live="polite"><p>{t(stages[job.state]??job.state)}</p><p>{job.current}</p>{job.total>0&&<progress className="retroarch-progress" max={job.total} value={job.bytes} aria-label={t('Progreso de descarga')}/>}<p>{job.completed} / {job.count} {t('componentes')}</p></div>}
  {!busy&&!installing&&!plan&&<button data-nav onClick={()=>void preparePlan()}>{t('Revisar instalación de nuevo')}</button>}
 </>}
 {step===3&&<>
  <p>{t('Indica una carpeta por consola. No moveremos ni modificaremos tus juegos.')}</p>
  <button data-nav disabled={busy} onClick={()=>void useFolders()}>{t('Crear y usar carpetas de Gemu')}</button>
  <div className="tv-setting-list">{config.collections.map(item=><div key={item.systemId}><TVField label={selectedSystems.find(sys=>sys.id===item.systemId)?.name??item.systemId} value={item.directory} onChange={directory=>patchCollection(item.systemId,{directory})}/><button data-nav disabled={busy} onClick={()=>void pick('roms',directory=>patchCollection(item.systemId,{directory}))}>{t('Elegir carpeta')}</button><button data-nav role="switch" aria-checked={item.recursive} onClick={()=>patchCollection(item.systemId,{recursive:!item.recursive})}>{t('Incluir subcarpetas')}: {t(item.recursive?'Sí':'No')}</button></div>)}</div>
  <TVField label={t('Carpeta de BIOS')} value={config.systemDirectory} onChange={systemDirectory=>setConfig(current=>({...current,systemDirectory}))}/><button data-nav onClick={()=>void pick('bios',systemDirectory=>setConfig(current=>({...current,systemDirectory})))}>{t('Elegir carpeta de BIOS')}</button>
  <p>{t('PlayStation admite CUE, CHD, PBP y listas M3U. Revisaremos sus archivos asociados para evitar discos duplicados o incompletos.')}</p>
 </>}
 {step===4&&<div className="tv-setting-list">
  <button data-nav role="switch" aria-checked={config.fullscreen} onClick={()=>setConfig(current=>({...current,fullscreen:!current.fullscreen}))}>{t('Pantalla completa sin bordes')}: {t(config.fullscreen?'Sí':'No')}</button>
  <button data-nav role="switch" aria-checked={config.rewind} onClick={()=>setConfig(current=>({...current,rewind:!current.rewind}))}>{t('Rebobinado')}: {t(config.rewind?'Sí':'No')}</button><p>{t('El rebobinado depende del núcleo y puede aumentar el consumo de recursos.')}</p>
  <TVField label={t('Controlador del mando')} value={({auto:'Automático',xinput:'XInput · Xbox',sdl2:'SDL2 · PlayStation / Nintendo',dinput:'DirectInput'} as Record<string,string>)[config.joypadDriver]??'Automático'} options={['Automático','XInput · Xbox','SDL2 · PlayStation / Nintendo','DirectInput']} onChange={value=>setConfig(current=>({...current,joypadDriver:({'Automático':'auto','XInput · Xbox':'xinput','SDL2 · PlayStation / Nintendo':'sdl2','DirectInput':'dinput'} as Record<string,string>)[value]}))}/>
  <p>{t('RetroArch utiliza sus perfiles de autoconfiguración. Si tu mando no responde, prueba otro controlador y abre RetroArch para comprobarlo.')}</p>
  <p>{t('L3 + R3 o F1 abre el menú de RetroArch. F10 o View + Menu abre el menú de Gemu.')}</p>
  <p>{t('Las partidas y estados se guardan por jugador y juego. Los mapeos de Gemu no se exportan automáticamente a RetroArch.')}</p>
  <p>{t('Para conservar ajustes hechos dentro de RetroArch, utiliza Guardar configuración actual en su menú.')}</p>
  <button data-nav disabled={busy} onClick={()=>{setError('');void retroHost().OpenRetroArchMenu(profileID,{...config,language}).catch(fail);}}>{t('Abrir RetroArch para probar')}</button>
  <p>{t('Cierra RetroArch al terminar la prueba para poder iniciar un juego desde Gemu.')}</p>
 </div>}
 {step===5&&<>
  {busy&&<p role="status">{t(preview?'Guardando tu biblioteca…':'Revisando juegos, núcleos y BIOS…')}</p>}
  {preview&&<><h3>{preview.games.length} {t('juegos preparados para tu biblioteca')}</h3><p>{t('Se actualizarán los juegos de RetroArch de este jugador. Las demás fuentes se conservarán.')}</p>
   <div className="tv-setting-list">{preview.checks.map(check=><div key={check.systemId}><strong>{status?.systems.find(sys=>sys.id===check.systemId)?.name}</strong><p>{t(check.coreReady?'Núcleo listo':'Falta instalar el núcleo')}</p>{check.firmware.map(item=><p key={item.name}>{item.name} · {item.region} · {t(({verified:'Verificada',missing:'No encontrada',invalid:'Contenido no válido'} as Record<string,string>)[item.state]??item.state)}</p>)}</div>)}</div>
   {preview.checks.some(item=>item.firmware.length>0)&&<p>{t('Las BIOS de PlayStation se comprueban por región y checksum. El funcionamiento sin BIOS depende de la versión del núcleo y del juego; una BIOS ausente no se da por verificada.')}</p>}
   <ul className="source-results">{preview.games.slice(previewPage*20,(previewPage+1)*20).map(entry=><li data-nav tabIndex={0} key={entry.game.ID}><strong>{entry.game.Title}</strong><span>{entry.metadata.platform??entry.metadata.genres.join(' · ')}</span></li>)}</ul>
   {preview.games.length>20&&<div className="tv-subnav"><button data-nav disabled={!previewPage} onClick={()=>setPreviewPage(value=>value-1)}>{t('Anterior')}</button><span>{previewPage+1} / {Math.ceil(preview.games.length/20)}</span><button data-nav disabled={(previewPage+1)*20>=preview.games.length} onClick={()=>setPreviewPage(value=>value+1)}>{t('Siguiente')}</button></div>}
   {!!preview.diagnostics.length&&<details><summary data-nav>{t('Incidencias')} ({preview.diagnostics.length})</summary>{preview.diagnostics.map((message,index)=><p key={index}>{t(message)}</p>)}</details>}
   {preview.checks.some(check=>!check.coreReady)&&<button data-nav onClick={()=>void preparePlan()}>{t('Instalar núcleos que faltan')}</button>}
   {preview.checks.some(check=>check.firmware.some(item=>item.state==='invalid'))&&<p>{t('Corrige las BIOS marcadas como no válidas y vuelve a revisar.')}</p>}
  </>}
  {!busy&&!preview&&<button data-nav onClick={()=>void review()}>{t('Volver a revisar juegos')}</button>}
 </>}
 </WizardFrame>}
 </div>;
}
