import {useEffect,useState} from 'react';
import {credentialHost,type CredentialProvider} from '../../library';
import {ProviderSetupWizard} from './ProviderSetupWizard';
import {setupProviders} from './setup-providers';
import {TVRow,TVScreen} from './TVControls';
const providers:CredentialProvider[]=['steamgriddb','screenscraper','steam','igdb','minimax','openrouter','compatible'];
export function CredentialManager(){
 const [selected,setSelected]=useState<CredentialProvider>(),[values,setValues]=useState<Record<string,string>>({});
 const [saved,setSaved]=useState<Record<string,number>>({}),[error,setError]=useState('');
 const back=()=>{setValues({});setSelected(undefined);};
 useEffect(()=>{if(selected)return;let cancelled=false;setSaved({});setError('');void (async()=>{try{const host=credentialHost(),statuses=await Promise.all(providers.map(async id=>[id,(await host.GetCredentialStatus(id)).savedFields.length] as const));if(!cancelled)setSaved(Object.fromEntries(statuses));}catch(reason){if(!cancelled)setError(String(reason));}})();return()=>{cancelled=true;};},[selected]);
 if(selected)return <ProviderSetupWizard key={selected} provider={selected} values={values} onChange={setValues} onComplete={back} onCancel={back}/>;
 return <TVScreen title="Credenciales" subtitle="Elige una conexión para abrir su asistente."><div className="tv-setting-list provider-fields">{providers.map(id=><TVRow key={id} label={setupProviders[id].name} value={saved[id]===undefined?'Consultando…':saved[id]?saved[id]+' campos guardados':'Configurar acceso'} onClick={()=>{setValues({});setSelected(id);}}/>)}{error&&<p role="status">{error}</p>}</div></TVScreen>;
}
