import {useEffect,useRef,useState} from 'react';
import {Keyboard,MousePointer2,X} from 'lucide-react';
import {gamepadHint,inputMethod,useInputMethod} from '../../input-method';
import {useLocalization} from '../../localization';
import {isForeground,subscribeActivity} from '../../activity';
import {hasGamepadCapture,subscribeGamepad} from './gamepad';
import {stickVelocity,textField,updateField,VirtualInputHotkeys,type TextField} from './virtual-input';
import './virtual-input.css';

const letters=['1234567890','qwertyuiop','asdfghjklñ','zxcvbnm,.-'];
const symbols=['1234567890','@#$%&*()_+','!?:;/\\=\"\'','áéíóúü[]{}'];
const commands=['shift','symbols','space','delete','done'];

export function VirtualInput(){
 const {t}=useLocalization(),input=useInputMethod();
 const [field,setField]=useState<TextField|null>(null),[pointer,setPointer]=useState(false),[shown,setShown]=useState(true);
 const [upper,setUpper]=useState(false),[symbol,setSymbol]=useState(false),[selected,setSelected]=useState({row:1,col:0});
 const [,refresh]=useState(0);
 const panel=useRef<HTMLDivElement>(null),cursor=useRef<HTMLDivElement>(null);
 const runtime=useRef({field:null as TextField|null,pointer:false,upper:false,symbol:false,selected:{row:1,col:0}});
 runtime.current={field,pointer,upper,symbol,selected};
 const close=()=>{const target=runtime.current.field;runtime.current.field=null;setField(null);if(target?.isConnected)target.focus({preventScroll:true});};
 const activate=(key:string)=>{
  const state=runtime.current,target=state.field;if(!target||!textField(target))return close();
  if(key==='done')return close();
  if(key==='shift'){state.upper=!state.upper;setUpper(state.upper);return;}
  if(key==='symbols'){
   state.symbol=!state.symbol;setSymbol(state.symbol);
   const length=state.selected.row===4?commands.length:Array.from((state.symbol?symbols:letters)[state.selected.row]).length;
   state.selected={...state.selected,col:Math.min(state.selected.col,length-1)};setSelected(state.selected);return;
  }
  updateField(target,key==='space'?' ':key==='delete'?'':state.upper?key.toLocaleUpperCase():key,key==='delete');
  refresh(value=>value+1);
 };
 const callbacks=useRef({close,activate});callbacks.current={close,activate};
 useEffect(()=>{
  let deviceIndex:number|null=null,last=0,blockAction=false,dpad=false;
  const hotkeys=new VirtualInputHotkeys();
  let lastField:TextField|null=null;
  const point={x:window.innerWidth/2,y:window.innerHeight/2};
  let hovered:HTMLElement|null=null;
  const clearHover=()=>{hovered?.classList.remove('virtual-pointer-target');hovered=null;};
  const capture=()=>blockAction||!!runtime.current.field||runtime.current.pointer;
  const open=(target:TextField)=>{blockAction=true;setPointer(false);runtime.current.pointer=false;runtime.current.field=target;runtime.current.upper=false;runtime.current.symbol=target.type==='number';runtime.current.selected={row:target.type==='number'?0:1,col:0};setField(target);setUpper(false);setSymbol(runtime.current.symbol);setSelected(runtime.current.selected);};
  const focus=(event:FocusEvent)=>{
   if(panel.current?.contains(event.target as Node))return;
   if(textField(event.target as Element))lastField=event.target as TextField;
   if(runtime.current.field&&event.target!==runtime.current.field){runtime.current.field=null;setField(null);}
  };
  const key=(event:KeyboardEvent)=>{
   if(!runtime.current.field&&!runtime.current.pointer)return;
   if(event.key==='Escape'){event.preventDefault();event.stopImmediatePropagation();callbacks.current.close();setPointer(false);runtime.current.pointer=false;}
   else if(!panel.current?.contains(event.target as Node)){callbacks.current.close();setPointer(false);runtime.current.pointer=false;}
  };
  const pause=()=>{clearHover();hotkeys.reset();last=0;setShown(isForeground());if(!isForeground()){setPointer(false);runtime.current.pointer=false;callbacks.current.close();}};
  const stopActivity=subscribeActivity(pause);
  window.addEventListener('blur',pause);window.addEventListener('focus',pause);document.addEventListener('visibilitychange',pause);
  document.addEventListener('focusin',focus);window.addEventListener('keydown',key,true);
  const unsubscribe=subscribeGamepad({priority:-1,capture,action:action=>{
   if(hasGamepadCapture(capture)||blockAction)return;
   const state=runtime.current;
   if(action==='menu')return;
   if(action==='back'){callbacks.current.close();setPointer(false);runtime.current.pointer=false;return;}
   if(state.pointer){
    const target=document.activeElement,step=action==='right'||action==='down'?1:action==='left'||action==='up'?-1:0;
    if(!dpad)return;
    if(step&&target instanceof HTMLSelectElement&&!target.disabled){
     const options=Array.from(target.options).filter(option=>!option.disabled),index=options.findIndex(option=>option.value===target.value);
     const next=options[Math.max(0,Math.min(options.length-1,index+step))];
     if(next){target.value=next.value;target.dispatchEvent(new Event('change',{bubbles:true}));}
    }
    if(step&&target instanceof HTMLInputElement&&target.type==='range'&&!target.disabled){
     if(step>0)target.stepUp();else target.stepDown();
     target.dispatchEvent(new Event('input',{bubbles:true}));target.dispatchEvent(new Event('change',{bubbles:true}));
    }
    return;
   }
   if(!state.field)return;
   if(action==='confirm'){
    const rows=[...(state.symbol?symbols:letters).map(row=>Array.from(row)),commands];
    callbacks.current.activate(rows[state.selected.row][state.selected.col]);return;
   }
   const row=(state.selected.row+(action==='up'?-1:action==='down'?1:0)+5)%5;
   const length=row===4?commands.length:Array.from((state.symbol?symbols:letters)[row]).length;
   const col=action==='left'?(state.selected.col-1+length)%length:action==='right'?(state.selected.col+1)%length:Math.min(state.selected.col,length-1);
   runtime.current.selected={row,col};setSelected({row,col});
  },devices:pads=>{
   blockAction=false;
   setShown(isForeground()&&!hasGamepadCapture(capture));
   if(!isForeground()||hasGamepadCapture(capture)){clearHover();hotkeys.reset();last=0;return;}
   const pad=pads.find(p=>p.mapping==='standard'&&p.index===inputMethod().index)??pads.find(p=>p.mapping==='standard');
   if(!pad){clearHover();callbacks.current.close();setPointer(false);runtime.current.pointer=false;hotkeys.reset();deviceIndex=null;return;}
   const buttons=new Set(pad.buttons.flatMap((button,index)=>button.pressed?[index]:[]));
   dpad=[12,13,14,15].some(index=>buttons.has(index));
   if(deviceIndex!==pad.index){deviceIndex=pad.index;hotkeys.reset();last=0;}
   const axes=pad.axes.map(stickVelocity);
   const shortcut=hotkeys.sample(buttons);
   if(!shortcut.ready){blockAction=true;return;}
   const {pressed}=shortcut;
   blockAction=shortcut.suppress;
   if(shortcut.keyboard){
    if(runtime.current.field)callbacks.current.close();
    else {
     const target=textField(document.activeElement)?document.activeElement:lastField&&textField(lastField)&&lastField.isConnected?lastField:null;
     if(target)open(target);
    }
   }
   if(shortcut.suppress)return;
   if(shortcut.pointer){callbacks.current.close();blockAction=true;runtime.current.pointer=!runtime.current.pointer;setPointer(runtime.current.pointer);}
   const state=runtime.current;
   if(state.field){
    if(!state.field.isConnected||!textField(state.field)){callbacks.current.close();return;}
    if(pressed(2))callbacks.current.activate('delete');
    if(pressed(3))callbacks.current.activate('shift');
    if(pressed(4))callbacks.current.activate('symbols');
    if(pressed(5))callbacks.current.activate('space');
   }
   const elapsed=Math.min(32,last?performance.now()-last:0)/1000;last=performance.now();
   if(!state.pointer){clearHover();return;}
   point.x=Math.max(2,Math.min(window.innerWidth-2,point.x+(axes[0]??0)*900*elapsed));
   point.y=Math.max(2,Math.min(window.innerHeight-2,point.y+(axes[1]??0)*900*elapsed));
   if(cursor.current)cursor.current.style.transform=`translate(${point.x}px,${point.y}px)`;
   const target=document.elementFromPoint(point.x,point.y);
   const interactive=target?.closest<HTMLElement>('button,a,input,select,textarea,[role="button"]');
   if(hovered!==interactive){clearHover();if(interactive&&!interactive.closest('[inert]')&&!interactive.matches(':disabled')){hovered=interactive;hovered.classList.add('virtual-pointer-target');}}
   if(pressed(0)&&interactive&&!interactive.closest('[inert]')&&!interactive.matches(':disabled')){
    blockAction=true;
    interactive.focus({preventScroll:true});
    interactive.click();
   }
   if(pressed(2)&&target&&!target.closest('[inert]'))target.dispatchEvent(new MouseEvent('contextmenu',{bubbles:true,cancelable:true,clientX:point.x,clientY:point.y,button:2}));
   const scroll=(axes[3]??0)*900*elapsed;
   if(scroll){let container=target;while(container){const overflow=getComputedStyle(container).overflowY;if(/auto|scroll/.test(overflow)&&container.scrollHeight>container.clientHeight){container.scrollTop+=scroll;break;}container=container.parentElement;}if(!container)window.scrollBy(0,scroll);}
  }});
  return()=>{clearHover();unsubscribe();stopActivity();document.removeEventListener('focusin',focus);window.removeEventListener('keydown',key,true);window.removeEventListener('blur',pause);window.removeEventListener('focus',pause);document.removeEventListener('visibilitychange',pause);};
 },[]);
 useEffect(()=>{if(field){panel.current?.querySelector<HTMLButtonElement>(`[data-key="${selected.row}:${selected.col}"]`)?.focus({preventScroll:true});}},[field,selected]);
 if(!shown)return null;
 const rows=(symbol?symbols:letters).map(row=>Array.from(row));
 const labels:Record<string,string>={shift:t('Mayúsculas'),symbols:symbol?'ABC':'#+=',space:t('Espacio'),delete:t('Borrar'),done:t('Listo')};
 return <>
  {pointer&&<div className="virtual-cursor" ref={cursor} aria-hidden="true"><MousePointer2 size={28}/></div>}
  {!field&&input.kind==='gamepad'&&<div className="virtual-input-tools"><span className="virtual-keyboard-shortcut"><Keyboard size={16}/> Home + <span className="pad-key">{gamepadHint(input.family,'favorite')}</span> · {t('Teclado virtual')}</span><button className="virtual-input-toggle" onClick={()=>{setPointer(value=>!value);runtime.current.pointer=!runtime.current.pointer;}} aria-pressed={pointer}><MousePointer2 size={16}/><span>{t(pointer?'Stick izquierdo: mover · Stick derecho: desplazar':'R3 · Ratón virtual')}</span>{pointer&&<><span className="pad-key">{gamepadHint(input.family,'confirm')}</span>{t('Clic')}<span className="pad-key">{gamepadHint(input.family,'back')}</span>{t('Salir')}</>}</button></div>}
  {field&&<div className="virtual-keyboard" ref={panel} role="dialog" aria-modal="true" aria-label={t('Teclado virtual')} onKeyDown={event=>{
   if(event.key==='Tab'){event.preventDefault();const buttons=Array.from(panel.current!.querySelectorAll<HTMLButtonElement>('button'));const index=buttons.indexOf(document.activeElement as HTMLButtonElement);buttons[(index+(event.shiftKey?-1:1)+buttons.length)%buttons.length]?.focus();event.stopPropagation();}
  }}>
   <div className="virtual-keyboard-heading"><span><Keyboard size={19}/>{t('Teclado virtual')}</span><button aria-label={t('Cerrar')} onClick={close}><X size={20}/></button></div>
   <div className="virtual-keyboard-preview" aria-live={field.type==='password'?'off':'polite'}>{field.type==='password'?'•'.repeat(field.value.length):field.value||field.placeholder||t('Escribe con el mando')}</div>
   <div className="virtual-keyboard-rows">{rows.map((row,r)=><div className="virtual-keyboard-row" key={r}>{row.map((value,c)=><button data-key={`${r}:${c}`} key={value} className={selected.row===r&&selected.col===c?'selected':''} onClick={()=>activate(value)}>{upper?value.toLocaleUpperCase():value}</button>)}</div>)}
   <div className="virtual-keyboard-row virtual-keyboard-actions">{commands.map((command,c)=><button data-key={`4:${c}`} key={command} className={(selected.row===4&&selected.col===c?'selected ':'')+(command==='done'?'virtual-keyboard-done':'')} aria-pressed={command==='shift'?upper:command==='symbols'?symbol:undefined} onClick={()=>activate(command)}>{labels[command]}</button>)}</div></div>
   <div className="virtual-keyboard-help"><span>{t('Cruceta / stick: elegir')}</span> · <span className="pad-key">{gamepadHint(input.family,'confirm')}</span> {t('Escribir')} · <span className="pad-key">{gamepadHint(input.family,'back')}</span> {t('Listo')}<span>{t('LB: símbolos · RB: espacio')} · Home + {gamepadHint(input.family,'favorite')}: {t('Listo')}</span></div>
  </div>}
 </>;
}
