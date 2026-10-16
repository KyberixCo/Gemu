import {preferencesStorage} from '../../persistence';
import {useEffect,useMemo,useState} from 'react';
import {useLocalization} from '../../localization';
import {libraryHost} from '../../library';
import type {GemuProfile,ProfileHost} from '../../profiles';
import {TVField} from '../management/TVControls';
import {ControllerProviders,nativeControllerProvider,type ControllerDevice} from './providers';
import {controllerAssociations,controllerRegistryKey,newRegisteredController,readControllerRegistry,type RegisteredController} from './registry';

export function ControllerDetails({device,onIdentity,onName,notify}:{device:ControllerDevice|null;onIdentity:(uuid:string|undefined)=>void;onName:(name:string)=>void;notify:(message:string)=>void}){
 const {t}=useLocalization();
 const [records,setRecords]=useState<RegisteredController[]>(()=>{try{return readControllerRegistry(JSON.parse(preferencesStorage.getItem(controllerRegistryKey)||'[]'));}catch{return [];}});
 const [players,setPlayers]=useState<GemuProfile[]>([]),[native,setNative]=useState<ControllerDevice[]>([]),[source,setSource]=useState(''),[error,setError]=useState('');
 const [revision,setRevision]=useState(0);
 const providers=useMemo(()=>{const service=new ControllerProviders();service.register(nativeControllerProvider());return service;},[]);
 useEffect(()=>{let cancelled=false;try{void (libraryHost() as ReturnType<typeof libraryHost>&ProfileHost).GetProfiles().then(items=>{if(!cancelled)setPlayers(items);}).catch(()=>{});}catch{/* Browser preview has no player host. */}return()=>{cancelled=true;};},[]);
 useEffect(()=>{let cancelled=false;let timer:ReturnType<typeof setTimeout>;
  const poll=async()=>{if(!document.hidden&&document.hasFocus()){const result=await providers.devices();if(cancelled)return;setNative(result.devices);setError(result.errors.length?t('No se pudo leer la información nativa del mando.'):'');}if(!cancelled)timer=setTimeout(()=>void poll(),2000);};
  void poll();return()=>{cancelled=true;clearTimeout(timer);};
 },[providers,t]);
 const record=device?controllerAssociations.resolve(device,records):undefined;
 useEffect(()=>{onIdentity(record?.uuid);},[record?.uuid,onIdentity,revision]);
 useEffect(()=>{onName(record?.name??'');},[record?.name,onName]);
 useEffect(()=>{setSource('');},[device?.id]);
 const persist=async(next:RegisteredController[])=>{setRecords(next);try{await preferencesStorage.setItem(controllerRegistryKey,JSON.stringify(next));}catch{notify(t('No se pudo guardar la identidad del mando.'));}};
 const edit=(patch:Partial<RegisteredController>)=>{if(record)persist(records.map(item=>item.uuid===record.uuid?{...item,...patch}:item));};
 const telemetry=native.find(item=>`${item.provider}:${item.id}`===source);
 const battery=telemetry?.battery;
 const batteryLabel=!telemetry?t('No disponible'):battery?.state==='wired'?t('Conectado por cable'):battery?.percent!==undefined?`${battery.percent}%`:t(({empty:'Vacía',low:'Baja',medium:'Media',full:'Completa'} as Record<string,string>)[battery?.level??'']??'Desconocida');
 return <div className="tv-setting-list">
  <p className="manager-hint">{t('El UUID de Gemu identifica un mando guardado. Si el dispositivo no ofrece identidad persistente, debes reasociarlo al reconectar.')}</p>
  {!record&&<button data-nav disabled={!device} className="manager-secondary" onClick={async()=>{if(!device)return;const item=newRegisteredController(device);try{await preferencesStorage.setItem(controllerRegistryKey,JSON.stringify([...records,item]));controllerAssociations.attach(device,item);setRecords([...records,item]);setRevision(value=>value+1);}catch{notify(t('No se pudo guardar la identidad del mando.'));}}}>{t('Registrar este mando')}</button>}
  <TVField label={t('Mando guardado')} disabled={!device||!records.length} value={record?`${record.name} · ${record.uuid}`:t('Sin asociar')} options={records.map(item=>`${item.name} · ${item.uuid}`)} onChange={value=>{const item=records.find(item=>`${item.name} · ${item.uuid}`===value);if(device&&item){try{controllerAssociations.attach(device,item);setRevision(count=>count+1);}catch(e){notify(t(String(e instanceof Error?e.message:e)));}}}}/>
  {record&&<><p className="manager-hint">UUID: {record.uuid}</p><TVField label={t('Nombre del mando')} value={record.name} onChange={name=>{const next=name.trim();if(next&&next.length<=64)edit({name:next});else notify(t('Usa un nombre de 1 a 64 caracteres.'));}}/><TVField label={t('Jugador asociado')} value={players.find(player=>player.id===record.playerId)?.name??t('Sin asociar')} options={[t('Sin asociar'),...players.map(player=>`${player.name} · ${player.id}`)]} onChange={value=>edit({playerId:players.find(player=>`${player.name} · ${player.id}`===value)?.id??''})}/></>}
  <p className="manager-hint">{t('XInput y el navegador no comparten identificadores. Selecciona manualmente el canal de batería; esta asociación dura mientras el canal siga conectado.')}</p>
  <TVField label={t('Canal de batería')} value={telemetry?.name??t('Sin asociar')} options={[t('Sin asociar'),...native.map(item=>`${item.name} · ${item.id}`)]} onChange={value=>setSource(native.find(item=>`${item.name} · ${item.id}`===value)?`${native.find(item=>`${item.name} · ${item.id}`===value)!.provider}:${native.find(item=>`${item.name} · ${item.id}`===value)!.id}`:'')}/>
  <p role="status">{t('Batería')}: {batteryLabel}{telemetry?.playerNumber?` · ${t('Número de mando')}: ${telemetry.playerNumber}`:''}</p>
  {error&&<p role="alert">{error}</p>}
  {telemetry?.capabilities['player-led']?.supported?<TVField label={t('LED de jugador')} value={String(telemetry.playerNumber??1)} options={['1','2','3','4']} onChange={value=>{void providers.setFeature(telemetry,'player-led',Number(value)).catch(()=>notify(t('No se pudo cambiar el LED del mando.')));}}/>:<p className="manager-hint">{t('LED de jugador: no disponible con el proveedor actual.')}</p>}
 </div>;
}
