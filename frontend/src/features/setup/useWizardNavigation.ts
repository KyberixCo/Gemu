import {useEffect} from 'react';
import {moveFocus,type FocusDirection} from '../../navigation';
import {subscribeGamepad} from '../feedback/gamepad';
import {navigationKey} from '../../keyboard-bindings';
export function useWizardNavigation(root:React.RefObject<HTMLElement|null>,enabled=true){
 useEffect(()=>{
  if(!enabled)return;
  const action=(name:string)=>{
   const scope=document.querySelector('.tv-editor')??root.current;if(!scope)return;
   if(name==='back'){window.dispatchEvent(new Event('gemu-management-back',{cancelable:true}));return;}
   if(name==='confirm'){
    if(document.activeElement instanceof HTMLInputElement&&scope.classList.contains('tv-editor'))scope.querySelector<HTMLElement>('button[data-nav]:not(:disabled)')?.click();
    else (document.activeElement as HTMLElement)?.click();return;
   }
   moveFocus(name as FocusDirection,scope);
  };
  const key=(event:KeyboardEvent)=>{
   const target=event.target as HTMLElement;
   const value=navigationKey(event);
   if(value==='Escape'){event.preventDefault();action('back');}
   else if(value==='Tab'){
    const scope=document.querySelector('.tv-editor')??root.current;if(!scope)return;
    const nodes=[...scope.querySelectorAll<HTMLElement>('[data-nav]:not(:disabled)')].filter(node=>node.getClientRects().length>0),index=nodes.indexOf(document.activeElement as HTMLElement);
    if(!nodes.length)return;event.preventDefault();nodes[(index+(event.shiftKey?-1:1)+nodes.length)%nodes.length]?.focus();
   }else if(value.startsWith('Arrow')&&!target.matches('input,textarea,select')){event.preventDefault();action(value.slice(5).toLowerCase());}
   else if((value==='Enter'||value===' ')&&!target.matches('input,textarea,select')){event.preventDefault();if(!event.repeat)action('confirm');}
  };
  const unsubscribe=subscribeGamepad({action});
  window.addEventListener('keydown',key);
  return()=>{window.removeEventListener('keydown',key);unsubscribe();};
 },[root,enabled]);
}
