import {useInputMethod,gamepadHint} from './input-method';
import {keyLabel,useKeyboardBindings} from './keyboard-bindings';
import {useLocalization} from './localization';
export function ActionHint({action,axis='horizontal'}:{action:'confirm'|'back'|'navigate'|'search';axis?:'horizontal'|'vertical'}){
 const input=useInputMethod(),{t}=useLocalization();
 const keys=useKeyboardBindings();
 if(action==='search')return input.kind==='keyboard'?<kbd>{t(keyLabel(keys.search))}</kbd>:null;
 if(input.kind==='gamepad')return <span className="pad-key">{gamepadHint(input.family,action)}</span>;
 if(action==='confirm'&&input.kind==='mouse')return <kbd>{t('Clic')}</kbd>;
 return <kbd>{t(action==='navigate'?`${keyLabel(keys[axis==='vertical'?'up':'left'])} ${keyLabel(keys[axis==='vertical'?'down':'right'])}`:keyLabel(keys[action]))}</kbd>;
}
