import {useEffect, useRef, useState, type ReactNode} from 'react';
import {ArrowLeft, ChevronRight} from 'lucide-react';

// Nested screens handle Back during capture; the management shell handles it afterward.
export function useTVBack(back:()=>void,enabled=true,priority=false) {
 const callback=useRef(back);callback.current=back;
 useEffect(()=>{if(!enabled)return;const handler=(event:Event)=>{if(event.defaultPrevented)return;event.preventDefault();callback.current();};window.addEventListener('gemu-management-back',handler,{capture:priority});return()=>window.removeEventListener('gemu-management-back',handler,{capture:priority});},[enabled,priority]);
}
export function TVScreen({title,subtitle,children}:{title:string;subtitle?:string;children:ReactNode}) {
 const root=useRef<HTMLDivElement>(null);
 useEffect(()=>{root.current?.querySelector<HTMLElement>('[data-nav]:not(:disabled)')?.focus({preventScroll:true});},[title]);
 return <div className="tv-screen" ref={root}><div className="tv-screen-heading" key={title}><h2>{title}</h2>{subtitle&&<p>{subtitle}</p>}</div>{children}</div>;
}
export function TVRow({label,value,onClick}:{label:string;value:string;onClick:()=>void}) {
 return <button data-nav className="tv-row" onClick={onClick}><strong>{label}</strong><span>{value || 'Sin configurar'}</span><ChevronRight size={22}/></button>;
}
export function TVField({label,value,onChange,secret=false,multiline=false,options}:{label:string;value:string;onChange:(value:string)=>void;secret?:boolean;multiline?:boolean;options?:readonly string[]}) {
 const [editing,setEditing]=useState(false),[draft,setDraft]=useState(value),[optionPage,setOptionPage]=useState(0);
 const trigger=useRef<HTMLButtonElement>(null),panel=useRef<HTMLDivElement>(null);
 const close=()=>{setEditing(false);requestAnimationFrame(()=>trigger.current?.focus({preventScroll:true}));};
 useEffect(()=>{if(!editing)return;panel.current?.querySelector<HTMLElement>('[data-nav]')?.focus({preventScroll:true});const handler=(event:Event)=>{event.preventDefault();event.stopImmediatePropagation();close();};window.addEventListener('gemu-management-back',handler,{capture:true});return()=>window.removeEventListener('gemu-management-back',handler,{capture:true});},[editing]);
 return <div className="tv-field"><button ref={trigger} data-nav className="tv-row" onClick={()=>{setDraft(value);setOptionPage(0);setEditing(true);}}><strong>{label}</strong><span>{secret?(value?'Credencial en sesión':'Sin credencial'):value || 'Sin configurar'}</span><ChevronRight size={22}/></button>{editing&&<div className="tv-editor" role="dialog" aria-modal="true" aria-label={label} ref={panel} onKeyDown={e=>{if(e.key==='Tab'){const nodes=[...panel.current!.querySelectorAll<HTMLElement>('[data-nav]:not(:disabled)')];const index=nodes.indexOf(document.activeElement as HTMLElement);e.preventDefault();nodes[(index+(e.shiftKey?-1:1)+nodes.length)%nodes.length]?.focus();}}}><div className="tv-screen-heading"><h2>{label}</h2><p>{options?'Elige con el mando y confirma.':'Introduce el valor con el teclado. Solo se modifica este ajuste.'}</p></div>{options?<div className="tv-options">{options.slice(optionPage*6,optionPage*6+6).map(option=><button data-nav key={option} aria-pressed={value===option} onClick={()=>{onChange(option);close();}}>{option}</button>)}{options.length>6&&<div className="tv-option-pages"><button data-nav disabled={optionPage===0} onClick={()=>setOptionPage(optionPage-1)}>Anterior</button><span>{optionPage+1} / {Math.ceil(options.length/6)}</span><button data-nav disabled={(optionPage+1)*6>=options.length} onClick={()=>setOptionPage(optionPage+1)}>Siguiente</button></div>}</div>:<><label className="manager-field"><span>{label}</span>{multiline?<textarea data-nav rows={4} value={draft} onChange={e=>setDraft(e.target.value)}/>:<input data-nav autoComplete="off" type={secret?'password':'text'} value={draft} onChange={e=>setDraft(e.target.value)}/>}</label><button data-nav className="manager-secondary" onClick={()=>{onChange(draft);close();}}>Aceptar al borrador</button></>}<button data-nav className="manager-secondary" onClick={close}><ArrowLeft size={18}/> Cancelar</button></div>}</div>;
}
