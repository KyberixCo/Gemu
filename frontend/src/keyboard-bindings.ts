import {profileKey} from './profiles';
import {useEffect,useState} from 'react';
export const keyboardActions=['confirm','back','up','down','left','right','search'] as const;
export type KeyboardAction=typeof keyboardActions[number];
export type KeyboardBindings=Record<KeyboardAction,string>;
export const keyboardDefaults:KeyboardBindings={confirm:'Enter',back:'Escape',up:'ArrowUp',down:'ArrowDown',left:'ArrowLeft',right:'ArrowRight',search:'Slash'};
export const keyboardOptions=['Enter','Escape','Space','ArrowUp','ArrowDown','ArrowLeft','ArrowRight','Slash',...['E','Q','W','A','S','D','Z','X','C','F','R'].map(key=>'Key'+key)];
export function keyLabel(code:string){return ({Enter:'Enter',Escape:'Esc',Space:'Espacio',ArrowUp:'↑',ArrowDown:'↓',ArrowLeft:'←',ArrowRight:'→',Slash:'/'} as Record<string,string>)[code]??code.replace(/^Key/,'');}
export function validKeyboardBindings(value:unknown):value is KeyboardBindings {if(!value||typeof value!=='object')return false;const bindings=value as KeyboardBindings;return keyboardActions.every(action=>keyboardOptions.includes(bindings[action]))&&new Set(keyboardActions.map(action=>bindings[action])).size===keyboardActions.length;}
export function readKeyboardBindings():KeyboardBindings {try{const saved=JSON.parse(localStorage.getItem(profileKey('keyboard.v1'))??'null');return validKeyboardBindings(saved)?saved:{...keyboardDefaults};}catch{return {...keyboardDefaults};}}
export function useKeyboardBindings(){const [bindings,setBindings]=useState(readKeyboardBindings);useEffect(()=>{const refresh=()=>setBindings(readKeyboardBindings());window.addEventListener('gemu-keyboard-bindings',refresh);window.addEventListener('storage',refresh);return()=>{window.removeEventListener('gemu-keyboard-bindings',refresh);window.removeEventListener('storage',refresh);};},[]);return bindings;}
export function assignKey(bindings:KeyboardBindings,action:KeyboardAction,code:string):KeyboardBindings {if(!keyboardOptions.includes(code))return bindings;const next={...bindings,[action]:code};const other=keyboardActions.find(item=>item!==action&&bindings[item]===code);if(other)next[other]=bindings[action];return next;}
export function navigationKey(event:KeyboardEvent):string {
 if(event.altKey||event.ctrlKey||event.metaKey)return '';
 const target=event.target as HTMLElement|null;if(target?.isContentEditable||['INPUT','TEXTAREA','SELECT'].includes(target?.tagName??''))return event.key;
 const code=event.code|| (event.key===' '?'Space':event.key==='/'?'Slash':event.key.length===1?'Key'+event.key.toUpperCase():event.key);
 const bindings=readKeyboardBindings();const action=keyboardActions.find(action=>bindings[action]===code);
 if(action)return ({confirm:'Enter',back:'Escape',up:'ArrowUp',down:'ArrowDown',left:'ArrowLeft',right:'ArrowRight',search:'/'} as Record<KeyboardAction,string>)[action];
 if(code==='Space'&&!Object.values(bindings).includes('Space'))return ' ';
 if(Object.values(keyboardDefaults).includes(code)){event.preventDefault();return '';}
 return event.key;
}
