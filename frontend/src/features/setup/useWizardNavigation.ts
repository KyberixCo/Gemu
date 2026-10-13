import {useEffect} from 'react';
import {moveFocus,type FocusDirection} from '../../navigation';
import {GamepadRepeater} from '../feedback/input';
export function useWizardNavigation(root:React.RefObject<HTMLElement|null>,enabled=true){
 useEffect(()=>{
  if(!enabled)return;
  const repeater=new GamepadRepeater();let frame=0;
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
   if(event.key==='Escape'){event.preventDefault();action('back');}
   else if(event.key==='Tab'){
    const scope=document.querySelector('.tv-editor')??root.current;if(!scope)return;
    const nodes=[...scope.querySelectorAll<HTMLElement>('[data-nav]:not(:disabled)')].filter(node=>node.getClientRects().length>0),index=nodes.indexOf(document.activeElement as HTMLElement);
    if(!nodes.length)return;event.preventDefault();nodes[(index+(event.shiftKey?-1:1)+nodes.length)%nodes.length]?.focus();
   }else if(event.key.startsWith('Arrow')&&!target.matches('input,textarea,select')){event.preventDefault();action(event.key.slice(5).toLowerCase());}
   else if((event.key==='Enter'||event.key===' ')&&!target.matches('input,textarea,select')){event.preventDefault();if(!event.repeat)action('confirm');}
  };
  const poll=(now:number)=>{
   const pad=navigator.getGamepads?.().find(pad=>pad?.connected);
   if(pad&&document.hasFocus()&&!document.hidden){const next=pad.buttons[0]?.pressed?'confirm':pad.buttons[1]?.pressed?'back':pad.buttons[12]?.pressed||pad.axes[1]<-.5?'up':pad.buttons[13]?.pressed||pad.axes[1]>.5?'down':pad.buttons[14]?.pressed||pad.axes[0]<-.5?'left':pad.buttons[15]?.pressed||pad.axes[0]>.5?'right':'';if(repeater.accept(next,now))action(next);}else repeater.reset();frame=requestAnimationFrame(poll);
  };
  window.addEventListener('keydown',key);frame=requestAnimationFrame(poll);
  return()=>{window.removeEventListener('keydown',key);cancelAnimationFrame(frame);};
 },[root,enabled]);
}
