import {useEffect,useRef,useState} from 'react';
import {contentHost,credentialHost} from '../../library';
import {CredentialEditor} from './CredentialEditor';
import {providerActivation,providerReadiness,providerSteps,setupProviders,sameCredentialValues,type SetupProvider} from './setup-providers';
import {TVField,TVRow,useTVBack} from './TVControls';
import {WizardFrame} from './WizardFrame';
export type ProviderSetupResult={enabled:boolean;priority:string;verified:boolean};
export function ProviderSetupWizard({provider,preferences=false,initialEnabled=false,initialPriority='1',values,onChange,onComplete,onCancel}:{provider:SetupProvider;preferences?:boolean;initialEnabled?:boolean;initialPriority?:string;values:Record<string,string>;onChange:(values:Record<string,string>)=>void;onComplete:(result:ProviderSetupResult)=>void;onCancel:()=>void}){
 const definition=setupProviders[provider],steps=providerSteps(provider,preferences);
 const [step,setStep]=useState(0),[saved,setSaved]=useState<string[]>([]),[checking,setChecking]=useState(false),[finishing,setFinishing]=useState(false),[loading,setLoading]=useState(!!definition.fields.length),[verified,setVerified]=useState(false),[status,setStatus]=useState(''),[enabled,setEnabled]=useState(initialEnabled),[priority,setPriority]=useState(initialPriority);
 const sequence=useRef(0),lastValues=useRef(values),alive=useRef(true);
 const refresh=async()=>{if(provider==='manual')return;const result=await credentialHost().GetCredentialStatus(provider);if(alive.current)setSaved(result.savedFields);};
 useEffect(()=>{alive.current=true;void refresh().catch(reason=>{if(alive.current)setStatus(String(reason));}).finally(()=>{if(alive.current)setLoading(false);});return()=>{alive.current=false;sequence.current++;};},[provider]);
 useEffect(()=>{if(!sameCredentialValues(lastValues.current,values)){setVerified(false);setStatus('');}lastValues.current=values;},[values]);
 const readiness=providerReadiness(provider,values,saved),current=steps[step];
 const willEnable=preferences&&providerActivation(provider,enabled,verified);
 const back=()=>{sequence.current++;setChecking(false);if(step)setStep(step-1);else onCancel();};
 useTVBack(()=>{if(!finishing&&!document.querySelector('.tv-editor'))back();},true,true);
 const check=async()=>{
  if(provider!=='screenscraper'&&provider!=='steamgriddb'||!readiness.ready)return;
  const id=++sequence.current;setChecking(true);setVerified(false);setStatus('Comprobando credenciales y acceso al servicio…');
  try{await contentHost().CheckContentProvider({id:provider,priority:Number(priority),credentials:values});if(alive.current&&id===sequence.current){setVerified(true);setEnabled(true);setStatus('Acceso confirmado. Todavía no se han importado juegos ni arte.');}}
  catch(reason){if(alive.current&&id===sequence.current)setStatus(String(reason));}
  finally{if(alive.current&&id===sequence.current)setChecking(false);}
 };
 const finish=async()=>{
  setFinishing(true);setStatus('');
  try{if(provider!=='manual'&&Object.values(values).some(Boolean)){await credentialHost().SaveProviderCredentials({provider,fields:values});if(!alive.current)return;onChange({});}
   onComplete({enabled:willEnable,priority,verified});
  }catch(reason){if(alive.current)setStatus(String(reason));}
  finally{if(alive.current)setFinishing(false);}
 };
 const fields=definition.fields;
 const panelValues=Object.fromEntries(fields.map(field=>[field.name,values[field.name]??'']));
 const setPanelValues=(next:Record<string,string>)=>{const other=Object.fromEntries(Object.entries(values).filter(([name])=>!fields.some(field=>field.name===name)));onChange({...other,...next});};
 return <WizardFrame title={`Configurar ${definition.name}`} steps={steps} current={step} actions={<><button data-nav className="manager-secondary" disabled={finishing} onClick={back}>{step?'Atrás':'Cancelar'}</button>{step<steps.length-1?<button data-nav className="primary" disabled={checking||loading||finishing} onClick={()=>{setStatus('');setStep(step+1);}}>Continuar</button>:<button data-nav className="primary" disabled={finishing||loading} onClick={()=>void finish()}>{finishing?'Guardando…':willEnable?'Guardar y activar':'Guardar y finalizar'}</button>}</>}>
 {current==='Introducción'&&<><p>{definition.purpose}</p><p>Te guiaremos por el acceso, la comprobación y las preferencias. Puedes guardar datos incompletos y terminar la configuración más adelante.</p>{provider==='screenscraper'&&<p>Introduce únicamente tu cuenta personal. Si esta versión de Gemu aún no tiene acceso a ScreenScraper, la comprobación lo indicará y podrás dejar el proveedor pendiente.</p>}{definition.docs&&<a data-nav className="manager-secondary" href={definition.docs} target="_blank" rel="noreferrer">Abrir documentación y acceso</a>}</>}
 {['Tu cuenta','Credenciales'].includes(current)&&<>
 {current==='Tu cuenta'&&<p>Introduce tu usuario y contraseña de ScreenScraper. Se comparten entre metadatos y arte. La cuenta es opcional, pero puede ser necesaria si el servicio restringe el acceso anónimo.</p>}
 {provider!=='manual'&&<CredentialEditor provider={provider} fields={fields} values={panelValues} onChange={setPanelValues} onStoredChange={()=>{setVerified(false);void refresh().catch(reason=>setStatus(String(reason)));}}/>}</>}
 {current==='Comprobación'&&<><p>Este paso comprueba el acceso antes de habilitar el proveedor.</p>{!readiness.ready&&<p role="status">{readiness.missing.length?'Falta completar o guardar: '+readiness.missing.map(name=>definition.fields.find(field=>field.name===name)?.label).join(', '):'Introduce ambos datos de tu cuenta, o deja los dos vacíos.'}</p>}<button data-nav className="primary" disabled={!readiness.ready||checking||loading} onClick={()=>void check()}>{checking?'Comprobando…':'Probar acceso'}</button><p>Puedes continuar sin verificar; el proveedor quedará desactivado y pendiente.</p></>}
 {current==='Preferencias'&&<><TVRow label="Usar proveedor" value={enabled?'Activado al finalizar':'Dejar pendiente'} onClick={()=>setEnabled(!enabled)}/><TVField label="Prioridad · 1 primero" value={priority} options={['1','2','3','4']} onChange={setPriority}/>{definition.canCheck&&!verified&&<p>La activación requiere una comprobación correcta en este asistente.</p>}</>}
 {current==='Revisión'&&<><h3>{definition.name}</h3><p>{verified?'Acceso comprobado':definition.canCheck?'Acceso pendiente de comprobar':readiness.ready?'Credenciales preparadas; la consulta se realizará cuando la integración esté disponible.':'Credenciales incompletas. Puedes guardarlas y terminar la configuración más adelante.'}</p>{preferences&&<p>{willEnable?`Se activará con prioridad ${priority}.`:'Quedará desactivado. Puedes reabrir este asistente desde Gestión.'}</p>}<p>Al finalizar se guardan de forma segura los valores nuevos. La importación de arte se realiza al configurar o actualizar una biblioteca.</p></>}
 {status&&<p role="status">{status}</p>}
 </WizardFrame>;
}
