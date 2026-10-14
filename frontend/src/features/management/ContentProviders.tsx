import {useLocalization} from '../../localization';
import {useState} from 'react';
import {Database} from 'lucide-react';
import {TVScreen} from './TVControls';
import {contentProviders,type ContentProviderId,type ContentSettings} from './content-providers';
import {ProviderSetupWizard} from './ProviderSetupWizard';
export {contentProviders,defaultContentSettings,readContentSettings} from './content-providers';
export type {ContentProviderId,ContentSettings} from './content-providers';
export function ContentProviders({kind,settings,onChange,credentials,onCredentials}:{kind:'metadata'|'art';settings:ContentSettings;onChange:(next:ContentSettings)=>void;credentials:Record<string,string>;onCredentials:(next:Record<string,string>)=>void}){
 const {t}=useLocalization();
 const [selected,setSelected]=useState<ContentProviderId|null>(null);
 const ids=(Object.keys(contentProviders) as ContentProviderId[]).filter(id=>contentProviders[id].kind===kind||(kind==='art'&&id==='screenscraper'));
 if(selected){
  const names=selected==='screenscraper'?['username','password']:selected==='igdb'?['clientId','clientSecret']:['apiKey'];
  const values=Object.fromEntries(names.map((name,index)=>[name,credentials[selected+':'+index]??'']));
  return <ProviderSetupWizard key={selected} provider={selected} preferences initialEnabled={settings[selected].enabled} initialPriority={settings[selected].priority} values={values} onChange={next=>{const other=Object.fromEntries(Object.entries(credentials).filter(([key])=>!key.startsWith(selected+':')));onCredentials({...other,...Object.fromEntries(names.map((name,index)=>[selected+':'+index,next[name]??'']))});}} onCancel={()=>setSelected(null)} onComplete={result=>{onChange({...settings,[selected]:{enabled:result.enabled,priority:result.priority}});setSelected(null);}}/>;
 }
 return <TVScreen title={kind==='metadata'?t("Fuentes de metadatos"):t("Fuentes de arte")} subtitle={t("Elige un proveedor. Su asistente te guiará por la configuración.")}><div className="tv-choice-grid">{ids.map(id=><button data-nav className="tv-module" key={id} onClick={()=>setSelected(id)}><Database size={30}/><strong>{t(contentProviders[id].name)}</strong><span>{settings[id].enabled?t("Configurar o actualizar"):t("Configurar por primera vez")}</span></button>)}</div></TVScreen>;
}
