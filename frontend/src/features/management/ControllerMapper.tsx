import {useLocalization} from '../../localization';
import {useEffect,useRef,useState} from 'react';
import {pressedControls,subscribeGamepad} from '../feedback/gamepad';
import {Gamepad2, Save} from 'lucide-react';
import {actions,controls,defaultProfile,families,label,type Control,type ControllerProfiles,type Family} from './controller-profiles';
import './controller-mapper.css';
import {TVField, TVRow} from './TVControls';
import {detectControllerFamily,controlPosition} from './controller-devices';
import {ControllerArtwork} from './ControllerArtwork';
import {ControllerDetails} from '../controllers/ControllerDetails';
import {browserController,type ControllerDevice} from '../controllers/providers';
import {useInputMethod} from '../../input-method';
import {KeyboardMapper} from './KeyboardMapper';
import {ActionHint} from '../../ActionHint';
import {navigationKey} from '../../keyboard-bindings';
export function ControllerMapper({profiles,onChange,notify,target,onTarget}:{profiles:ControllerProfiles;onChange:(profiles:ControllerProfiles)=>void;notify:(message:string)=>void;target:string;onTarget:(target:string)=>void}) {
 const {t}=useLocalization();
 const input=useInputMethod();
 const [inputView,setInputView]=useState('Automático');
 const showKeyboard=inputView==='Teclado y ratón'||inputView==='Automático'&&input.kind!=='gamepad';
 const [page,setPage]=useState('Perfil');
 const [family,setFamily]=useState<Family>('xbox'),[selected,setSelected]=useState<Control>('south'),[testing,setTesting]=useState(false),[pressed,setPressed]=useState<Control|null>(null);
 const [device,setDevice]=useState(''),[physical,setPhysical]=useState<Control[]>([]);
 const [identity,setIdentity]=useState<string>(),[descriptor,setDescriptor]=useState<ControllerDevice|null>(null);
 const [customName,setCustomName]=useState('');
 const [devices,setDevices]=useState<{index:number;id:string;mapping:string}[]>([]);
 const [deviceIndex,setDeviceIndex]=useState<number|null>(null),[automatic,setAutomatic]=useState(true),[standard,setStandard]=useState(false);
 const deviceIndexRef=useRef<number|null>(null),automaticRef=useRef(true);
 deviceIndexRef.current=deviceIndex;automaticRef.current=automatic;
 const testingRef=useRef(false);testingRef.current=!showKeyboard&&testing&&page==='Asignación';
 useEffect(()=>{const key=(event:KeyboardEvent)=>{if(testingRef.current&&navigationKey(event)==='Escape'){event.preventDefault();event.stopImmediatePropagation();setTesting(false);}};window.addEventListener('keydown',key,true);return()=>window.removeEventListener('keydown',key,true);},[]);
 useEffect(()=>subscribeGamepad({capture:()=>testingRef.current,action:name=>{if(name==='back'&&testingRef.current)setTesting(false);},devices:(pads,active)=>{
  const nextDevices=pads.map(({index,id,mapping})=>({index,id,mapping}));
  setDevices(previous=>JSON.stringify(previous)===JSON.stringify(nextDevices)?previous:nextDevices);
  const chosen=deviceIndexRef.current===null?(active??pads[0]??null):(pads.find(pad=>pad.index===deviceIndexRef.current)??null);
  const pad=chosen&&document.hasFocus()&&!document.hidden?chosen:null;
  setDevice(pad?.id??'');
  const nextDescriptor=pad?browserController(pad.index,pad.id):null;
  setDescriptor(previous=>previous?.id===nextDescriptor?.id?previous:nextDescriptor);
  setStandard(pad?.mapping==='standard');
  const detected=pad?detectControllerFamily(pad.id):null;
  if(automaticRef.current&&detected)setFamily(detected);
  const next=pad?pressedControls(pad):[];
  setPhysical(previous=>previous.join(',')===next.join(',')?previous:next);
 }}),[]);
 const baseKey=target+':'+family,key=identity?baseKey+':'+identity:baseKey,profile=profiles[key]??profiles[baseKey]??defaultProfile();
 const update=(patch:Partial<typeof profile>)=>onChange({...profiles,[key]:{...profile,...patch}});
 const select=(id:Control)=>{setSelected(id);if(testing)setPressed(id);};
 return <div className="controller-mapper">
 <TVField label={t('Método de entrada')} value={inputView} options={['Automático','Teclado y ratón','Mando']} onChange={setInputView}/>
 {showKeyboard?<KeyboardMapper notify={notify}/>:<>
 <div className="mapper-device-status" role="status"><strong>{device?(detectControllerFamily(device)?t(families[detectControllerFamily(device)!]):t('Mando no identificado')):t('Sin mando conectado')}</strong><span>{device||t('Conecta un mando y pulsa un botón para detectarlo.')}</span>{device&&!standard&&<span>{t('Este dispositivo no ofrece mapeo estándar. La lectura de botones no está disponible.')}</span>}</div>
 <div className="tv-setting-list"><TVField label={t('Dispositivo')} value={deviceIndex===null?t('Automático'):devices.find(pad=>pad.index===deviceIndex)?`${deviceIndex+1} · ${devices.find(pad=>pad.index===deviceIndex)!.id}`:t('Dispositivo desconectado')} options={[t('Automático'),...devices.map(pad=>`${pad.index+1} · ${pad.id}`)]} onChange={value=>{setDeviceIndex(devices.find(pad=>`${pad.index+1} · ${pad.id}`===value)?.index??null);setPressed(null);}}/><TVRow label={t('Detectar familia automáticamente')} value={automatic?t('Activado'):t('Desactivado')} onClick={()=>setAutomatic(!automatic)}/></div>
 <div className="tv-subnav">{['Perfil','Identidad','Asignación','Calibración','Resumen'].map(name=><button data-nav key={name} aria-pressed={page===name} onClick={()=>setPage(name)}>{t(name)}</button>)}</div>
 <div hidden={page!=='Identidad'}><ControllerDetails device={descriptor} onIdentity={setIdentity} onName={setCustomName} notify={notify}/></div>
 {customName&&<p className="manager-hint">{customName}</p>}
 {identity&&<p className="manager-hint">{t('Mapeo de este mando')} · {identity}</p>}
 {page==='Perfil'&&<div className="tv-setting-list" key={page}><TVField label={t("Familia de mando")} value={families[family]} options={Object.values(families)} onChange={name=>{setAutomatic(false);setFamily((Object.keys(families) as Family[]).find(id=>families[id]===name)!);setPressed(null);}}/><TVField label={t("Destino")} value={target} options={['Global','RetroArch','Dolphin','PCSX2','PPSSPP']} onChange={name=>{onTarget(name);setPressed(null);}}/><TVRow label={t("Asignar botones")} value={profiles[key]?t("Perfil personalizado"):t("Plantilla base")} onClick={()=>setPage('Asignación')}/></div>}
 {page==='Calibración'&&<div className="tv-setting-list" key={page}><TVField label={t("Zona muerta · %")} value={String(profile.deadzone)} options={Array.from({length:11},(_,i)=>String(i*5))} onChange={value=>update({deadzone:Number(value)})}/><TVField label={t("Sensibilidad · %")} value={String(profile.sensitivity)} options={Array.from({length:21},(_,i)=>String(50+i*5))} onChange={value=>update({sensitivity:Number(value)})}/><TVRow label={t("Invertir eje Y")} value={profile.invertY?t("Activado"):t("Desactivado")} onClick={()=>update({invertY:!profile.invertY})}/><TVRow label={t("Restaurar este perfil")} value="Volver a la plantilla" onClick={()=>{update(defaultProfile());setPressed(null);}}/></div>}
 {page==='Asignación'&&<div className="mapper-layout"><div className="manager-card mapper-stage"><div className="manager-card-heading"><div><h3>{t(families[family])}</h3><p>{t("Selecciona una entrada del mando para cambiar su acción.")}</p></div><Gamepad2 size={25}/></div>
 <div className={'controller-diagram '+family} role="group" aria-label={t('Entradas de {family}',{family:t(families[family])})}>
 <ControllerArtwork family={family}/>
 {controls.map(c=>{const {x,y}=controlPosition(family,c.id);return <button data-nav key={c.id} className={'controller-input '+c.id+(selected===c.id?' selected':'')+(testing&&(physical.includes(c.id)||pressed===c.id)?' pressed':'')} style={{left:x+'%',top:y+'%'}} aria-label={t(label(family,c.id))+' · '+t(profile.bindings[c.id])} aria-pressed={selected===c.id} onClick={()=>select(c.id)}>{t(label(family,c.id))}</button>;})}
 </div>
 <div className="mapper-stage-footer"><span>{target} / {families[family].split(' · ')[0]}</span><button data-nav className="manager-secondary" aria-pressed={testing} onClick={()=>{setTesting(!testing);setPressed(null);}}>{testing?t("Salir de prueba visual"):t("Probar visualmente")}</button></div><p className="manager-hint">{testing?<>{t("Pulsa los botones del mando para ver su asignación.")} <ActionHint action="back"/> {t("Salir de la prueba")}</>:t("Plantilla editable; la familia seleccionada no implica que haya un dispositivo conectado.")}</p>{testing&&<div className="mapper-test" role="status">{device&&<p>{device}</p>}{physical.length?physical.map(id=>t(label(family,id))+' → '+t(profile.bindings[id])).join(' · '):pressed?t(label(family,pressed))+' → '+t(profile.bindings[pressed]):t("Conecta un mando compatible y pulsa un botón.")}</div>}
 </div><div className="manager-card mapper-editor"><span className="manager-badge">{t("ASIGNACIÓN")}</span><h3>{t("Entrada")} {t(label(family,selected))}</h3><p>{t("Botón físico → acción del perfil")}</p><TVField label={t('Acción en {target}',{target})} value={profile.bindings[selected]} options={actions} onChange={value=>update({bindings:{...profile.bindings,[selected]:value}})}/><p className="manager-hint">{t("Selecciona una entrada en el dibujo y después su acción. Guarda desde la barra inferior.")}</p><button data-nav className="manager-secondary" onClick={()=>notify(t('Aplicación a {target} pendiente del host.',{target}))}><Save size={18}/> {t("Aplicar al destino · demo")}</button></div></div>}

 {page==='Resumen'&&<div className="manager-card tv-binding-summary"><div className="manager-card-heading"><div><h3>{t("Resumen de asignaciones")}</h3><p>{family==='switch'?t("Los botones Nintendo respetan sus etiquetas B / A / Y / X; la plantilla usa posiciones físicas."):t("La plantilla usa posiciones físicas para mantener una distribución consistente.")}</p></div></div><div className="mapper-bindings">{controls.map(c=><button data-nav key={c.id} aria-pressed={selected===c.id} onClick={()=>{select(c.id);setPage('Asignación');}}><strong>{t(label(family,c.id))}</strong><span>{t(profile.bindings[c.id])}</span></button>)}</div><p className="manager-hint">{t("Sticks analógicos: movimiento / cámara previstos. Touchpad, giroscopio, vibración, gatillos adaptativos y Joy-Con individuales requieren integración nativa.")}</p></div>}
 </>}
 </div>;
}
