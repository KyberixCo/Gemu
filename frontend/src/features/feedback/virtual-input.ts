export type TextEdit={value:string;caret:number};
// Only buttons gate hotkey readiness. Extra axes (for example trigger axes
// resting at -1) must not prevent R3 from becoming available.
export class VirtualInputHotkeys{
 private ready=false;
 private previous=new Set<number>();
 private chord=false;
 reset(){this.ready=false;this.previous.clear();this.chord=false;}
 sample(buttons:ReadonlySet<number>){
  const previous=this.previous;
  const pressed=(index:number)=>buttons.has(index)&&!previous.has(index);
  this.previous=new Set(buttons);
  if(!this.ready){if(!buttons.size)this.ready=true;return {ready:false,keyboard:false,pointer:false,suppress:true,pressed};}
  if(!buttons.has(16)&&!buttons.has(2))this.chord=false;
  const keyboard=!this.chord&&buttons.has(16)&&buttons.has(2);
  if(keyboard)this.chord=true;
  return {ready:true,keyboard,pointer:pressed(11)&&!buttons.has(16)&&!this.chord,suppress:buttons.has(16)||this.chord,pressed};
 }
}
export function editText(value:string,start:number,end:number,text:string,maxLength=-1):TextEdit{
 const available=maxLength<0?Infinity:Math.max(0,maxLength-(value.length-(end-start)));
 const insertion=Array.from(text).reduce((result,char)=>result.length+char.length<=available?result+char:result,'');
 return {value:value.slice(0,start)+insertion+value.slice(end),caret:start+insertion.length};
}
export function deleteText(value:string,start:number,end:number):TextEdit{
 const previous=start===end?start-(Array.from(value.slice(0,start)).at(-1)?.length??0):start;
 return {value:value.slice(0,previous)+value.slice(end),caret:previous};
}
export function stickVelocity(axis:number){
 if(!Number.isFinite(axis)||Math.abs(axis)<=0.18)return 0;
 return Math.sign(axis)*Math.min(1,((Math.min(1,Math.abs(axis))-0.18)/0.82)**2);
}
export type TextField=HTMLInputElement|HTMLTextAreaElement;
export function textField(element:Element|null):element is TextField{
 return (element instanceof HTMLTextAreaElement||element instanceof HTMLInputElement&&['text','search','url','email','tel','password','number'].includes(element.type))&&!element.disabled&&!element.readOnly&&!element.closest('[inert]');
}
export function updateField(field:TextField,text:string,remove=false){
 const start=field.selectionStart??field.value.length,end=field.selectionEnd??start;
 const edit=remove?deleteText(field.value,start,end):editText(field.value,start,end,text,field.maxLength);
 if(edit.value===field.value)return;
 const prototype=field instanceof HTMLTextAreaElement?HTMLTextAreaElement.prototype:HTMLInputElement.prototype;
 Object.getOwnPropertyDescriptor(prototype,'value')?.set?.call(field,edit.value);
 // Use the native setter so React's value tracker observes the change.
 field.dispatchEvent(new Event('input',{bubbles:true}));
 try{field.setSelectionRange(edit.caret,edit.caret);}catch{/* Number/email inputs do not expose a selection. */}
}
