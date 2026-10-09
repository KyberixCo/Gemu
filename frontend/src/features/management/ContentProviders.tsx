import {useState} from 'react';
import { ExternalLink, Database } from 'lucide-react';
import {TVField, TVRow, TVScreen, useTVBack} from './TVControls';

export const contentProviders = {
 igdb: {name:'IGDB',kind:'metadata',description:'Fichas, géneros y datos de juegos.',docs:'https://api-docs.igdb.com/',credentials:['Client ID','Client secret']},
 screenscraper: {name:'ScreenScraper',kind:'metadata',description:'Información y medios para bibliotecas retro.',docs:'https://www.screenscraper.fr/webapi2.php',credentials:['ID de desarrollador','Contraseña de desarrollador','Usuario','Contraseña']},
 steamgriddb: {name:'SteamGridDB',kind:'art',description:'Portadas, fondos, iconos y logotipos.',docs:'https://www.steamgriddb.com/api/v2',credentials:['Clave API']},
 manual: {name:'Archivos propios',kind:'art',description:'Rutas de arte elegidas en cada ficha. Sin API.',docs:'',credentials:[]},
} as const;
export type ContentProviderId = keyof typeof contentProviders;
export type ContentSettings = Record<ContentProviderId,{enabled:boolean;priority:string}>;
export const defaultContentSettings = ():ContentSettings => Object.fromEntries(Object.keys(contentProviders).map((id,i)=>[id,{enabled:id==='manual',priority:String(i+1)}])) as ContentSettings;
export function readContentSettings(value:unknown):ContentSettings {
 const result=defaultContentSettings();
 if(value && typeof value==='object') for(const id of Object.keys(result) as ContentProviderId[]) {
  const entry=(value as Partial<ContentSettings>)[id];
  if(entry && typeof entry.enabled==='boolean' && /^[1-4]$/.test(entry.priority)) result[id]={enabled:entry.enabled,priority:entry.priority};
 }
 return result;
}
export function ContentProviders({kind,settings,onChange,credentials,onCredentials}:{kind:'metadata'|'art';settings:ContentSettings;onChange:(next:ContentSettings)=>void;credentials:Record<string,string>;onCredentials:(next:Record<string,string>)=>void}) {
 const [selected,setSelected]=useState<ContentProviderId|null>(null),[page,setPage]=useState('Preferencias');
 useTVBack(()=>{setSelected(null);setPage('Preferencias');},!!selected,true);
 const ids=(Object.keys(contentProviders) as ContentProviderId[]).filter(id=>contentProviders[id].kind===kind||(kind==='art'&&id==='screenscraper'));
 if(!selected)return <TVScreen title={kind==='metadata'?'Fuentes de metadatos':'Fuentes de arte'} subtitle="Elige un proveedor para configurar su uso y acceso."><div className="tv-choice-grid">{ids.map(id=><button data-nav className="tv-module" key={id} onClick={()=>setSelected(id)}><Database size={30}/><strong>{contentProviders[id].name}</strong><span>{settings[id].enabled?'Habilitado · sin verificar':'Desactivado'}</span></button>)}</div></TVScreen>;
 const provider=contentProviders[selected],setting=settings[selected];
 return <TVScreen title={provider.name} subtitle={provider.description}><div className="tv-subnav">{['Preferencias',...(provider.credentials.length?['Credenciales']:[])].map(name=><button data-nav key={name} aria-pressed={page===name} onClick={()=>setPage(name)}>{name}</button>)}</div><div className="tv-setting-list" key={page}>{page==='Preferencias'?<><TVRow label="Usar proveedor" value={setting.enabled?'Activado':'Desactivado'} onClick={()=>onChange({...settings,[selected]:{...setting,enabled:!setting.enabled}})}/><TVField label="Prioridad · 1 primero" value={setting.priority} options={['1','2','3','4']} onChange={priority=>onChange({...settings,[selected]:{...setting,priority}})}/>{provider.docs&&<a data-nav className="tv-row" href={provider.docs} target="_blank" rel="noreferrer"><strong>Documentación y acceso</strong><span>Abrir documentación oficial</span><ExternalLink size={20}/></a>}<p className="manager-hint">{provider.credentials.length?'Conexión pendiente del host. Habilitar no verifica las credenciales.':'Configura las rutas desde Metadatos → Arte.'}</p></>:<>{provider.credentials.map((label,i)=><TVField key={label} label={label} secret value={credentials[`${selected}:${i}`]??''} onChange={value=>onCredentials({...credentials,[`${selected}:${i}`]:value})}/>)}<div className="tv-credential-footer"><p className="manager-hint">Solo en esta sesión. No se guardan con la configuración.</p><button data-nav className="manager-secondary" onClick={()=>onCredentials(Object.fromEntries(Object.entries(credentials).filter(([key])=>!key.startsWith(`${selected}:`))))}>Borrar credenciales</button></div></>}</div></TVScreen>;
}
