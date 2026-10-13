import {useEffect,useRef,useState} from 'react';
import {credentialHost,type CredentialProvider,type CredentialStorage} from '../../library';
import {TVField} from './TVControls';

export function CredentialEditor({provider,fields,values,onChange,onStoredChange}:{provider:CredentialProvider;fields:{name:string;label:string}[];values:Record<string,string>;onChange:(values:Record<string,string>)=>void;onStoredChange?:()=>void}){
 const [saved,setSaved]=useState<string[]>([]),[busy,setBusy]=useState(false),[status,setStatus]=useState('');
 const [storage,setStorage]=useState<CredentialStorage>({name:'Almacén seguro',available:false});
 const sequence=useRef(0);
 useEffect(()=>{
  const id=++sequence.current;setSaved([]);setStorage({name:'Almacén seguro',available:false});setBusy(true);setStatus('');
  void (async()=>{try{const result=await credentialHost().GetCredentialStatus(provider);if(sequence.current===id){setSaved(result.savedFields);setStorage(result.storage??{name:'Almacén seguro',available:true});}}catch(reason){if(sequence.current===id)setStatus(String(reason));}finally{if(sequence.current===id)setBusy(false);}})();
  return()=>{sequence.current++;};
 },[provider]);
 const save=async()=>{
  const id=++sequence.current;setBusy(true);setStatus('');
  try{
   const host=credentialHost();await host.SaveProviderCredentials({provider,fields:values});
   if(sequence.current!==id)return;
   onChange({});onStoredChange?.();
   const result=await host.GetCredentialStatus(provider);
   if(sequence.current===id){setSaved(result.savedFields);setStatus('Credenciales guardadas en el almacén seguro.');}
  }catch(reason){if(sequence.current===id)setStatus(String(reason));}
  finally{if(sequence.current===id)setBusy(false);}
 };
 const remove=async()=>{
  const id=++sequence.current;setBusy(true);setStatus('');
  try{await credentialHost().DeleteProviderCredentials(provider);if(sequence.current===id){setSaved([]);onChange({});onStoredChange?.();setStatus('Credenciales eliminadas de este perfil.');}}
  catch(reason){if(sequence.current===id)setStatus(String(reason));}
  finally{if(sequence.current===id)setBusy(false);}
 };
 return <div className="credential-editor"><p className="manager-hint">{storage.available?`${storage.name} · separado por perfil. Los campos vacíos conservan lo guardado; los valores nuevos reemplazan solo ese campo al guardar.`:'No hay almacenamiento seguro disponible. Puedes usar valores temporales durante esta sesión.'}</p>
 {fields.map(field=><TVField key={field.name} label={field.label} secret stored={saved.includes(field.name)} disabled={busy} value={values[field.name]??''} onChange={value=>onChange({...values,[field.name]:value})}/>)}
 <div className="manager-actions"><button data-nav className="manager-secondary" disabled={busy||!storage.available||!Object.values(values).some(Boolean)} onClick={()=>void save()}>Guardar credenciales</button><button data-nav className="manager-secondary" disabled={busy||!storage.available||!saved.length} onClick={()=>void remove()}>Eliminar guardadas</button><button data-nav className="manager-secondary" disabled={busy||!Object.values(values).some(Boolean)} onClick={()=>onChange({})}>Descartar valores nuevos</button></div>
 <p role="status">{busy?'Procesando credenciales…':status||`${saved.length} campos guardados en este perfil.`}</p></div>;
}
