import {preferencesStorage} from '../../persistence';
import {useState} from 'react';
import {Keyboard,Mouse} from 'lucide-react';
import {keyboardActions,keyboardDefaults,keyboardOptions,keyLabel,readKeyboardBindings,assignKey,type KeyboardAction} from '../../keyboard-bindings';
import {profileKey} from '../../profiles';
import {useLocalization} from '../../localization';
import {TVField} from './TVControls';
const names:Record<KeyboardAction,string>={confirm:'Seleccionar',back:'Volver',up:'Arriba',down:'Abajo',left:'Izquierda',right:'Derecha',search:'Buscar'};
export function KeyboardMapper({notify}:{notify:(text:string)=>void}){
 const {t}=useLocalization(),[bindings,setBindings]=useState(readKeyboardBindings);
 const save=async(next:typeof bindings)=>{try{await preferencesStorage.setItem(profileKey('keyboard.v1'),JSON.stringify(next));setBindings(next);window.dispatchEvent(new Event('gemu-keyboard-bindings'));notify(t('Atajos de Gemu guardados.'));}catch{notify(t('No se pudieron guardar los atajos.'));}};
 return <div className="keyboard-mapper"><div className="mapper-device-status"><Keyboard/><strong>{t('Teclado y ratón')}</strong><span>{t('Navegación de Gemu')}</span></div><p className="manager-hint">{t('Estos atajos se aplican a Gemu. Si eliges una tecla ocupada, las asignaciones se intercambian.')}</p><div className="keyboard-binding-grid">{keyboardActions.map(action=><TVField key={action} label={t(names[action])} value={keyLabel(bindings[action])} options={keyboardOptions.map(keyLabel)} onChange={value=>{const code=keyboardOptions.find(code=>keyLabel(code)===value);if(code)save(assignKey(bindings,action,code));}}/>)}</div><div className="mapper-device-status"><Mouse/><span>{t('Clic izquierdo: seleccionar · Rueda: desplazar · Tab: cambiar foco')}</span></div><p className="manager-hint">{t('F10 abre el menú del juego. Los controles del juego se configuran dentro de cada juego.')}</p><button data-nav className="manager-secondary" onClick={()=>save({...keyboardDefaults})}>{t('Restaurar atajos')}</button></div>;
}
