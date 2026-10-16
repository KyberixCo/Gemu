// Shared lifecycle state: gameplay can start before the WebView loses focus.
let gameplay=false;
let assisting=false;
const listeners=new Set<()=>void>();
export function isGameplay(){return gameplay;}
export function setGameplay(value:boolean){
 if(gameplay===value)return;
 gameplay=value;
 document.documentElement.dataset.lowPower=String(value);
 for(const listener of listeners)listener();
}
export function subscribeActivity(listener:()=>void){listeners.add(listener);return()=>{listeners.delete(listener);};}
export function setAssisting(value:boolean){
 if(assisting===value)return;
 assisting=value;for(const listener of listeners)listener();
}
export function isForeground(){return !document.hidden&&document.hasFocus()&&!gameplay&&!assisting;}
