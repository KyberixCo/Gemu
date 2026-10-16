import {useEffect,useState,type ReactNode} from 'react';
import {hydratePreferences} from './persistence';
import {translate,resolveLanguage} from './localization';

export function PersistenceGate({children}:{children:ReactNode}){
 const [ready,setReady]=useState(false),[error,setError]=useState(''),[attempt,setAttempt]=useState(0);
 useEffect(()=>{let cancelled=false;setError('');void hydratePreferences().then(()=>{if(!cancelled)setReady(true);}).catch(reason=>{if(!cancelled)setError(String(reason));});return()=>{cancelled=true;};},[attempt]);
 if(ready)return <>{children}</>;
 const language=resolveLanguage(null,navigator.languages);
 return <section className="profile-gate"><div className="profile-card">{error?<><p role="alert">{translate(error,language)}</p><button onClick={()=>setAttempt(value=>value+1)}>{translate('Reintentar conexión',language)}</button></>:<p role="status">{translate('Cargando perfiles…',language)}</p>}</div></section>;
}
