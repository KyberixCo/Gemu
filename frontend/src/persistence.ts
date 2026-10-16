// Native reads use a hydrated memory cache; writes resolve only after SQLite
// commits. Browser previews retain localStorage without pretending to be native.
type PreferenceState={values:Record<string,string>;legacyImported:boolean};
type PreferenceHost={LoadPreferences:(profileID:string,legacy:Record<string,string>|null)=>Promise<PreferenceState>;SavePreference:(profileID:string,key:string,value:string)=>Promise<void>};
const globals=['gemu.language.v1','gemu.controllers.v1'];
const playerKeys=['gemu.favorites','gemu.feedback','gemu.management.v1','gemu.setup.v1','keyboard.v1'];
const cache=new Map<string,string>();
const ready=new Set<string>();
let writes:Promise<void>=Promise.resolve();
function host():PreferenceHost|undefined {
 return (globalThis as typeof globalThis&{window?:{go?:{main?:{App?:PreferenceHost}}}}).window?.go?.main?.App;
}
function owner(key:string){
 if(globals.includes(key))return {id:'',key};
 const match=/^gemu\.profile\.([^.]+)\.(.+)$/.exec(key);
 if(!match||!playerKeys.includes(match[2]))throw new Error('Invalid preference key');
 return {id:match[1],key:match[2]};
}
function fullKey(id:string,key:string){return id?`gemu.profile.${id}.${key}`:key;}
export async function hydratePreferences(id=''){
 const backend=host();if(!backend)return;
 if(typeof backend.LoadPreferences!=='function'||typeof backend.SavePreference!=='function')throw new Error('Reinicia Gemu con dev.ps1 o abre el ejecutable actualizado para usar SQLite.');
 await writes;
 let saved=await backend.LoadPreferences(id,null);
 if(!saved.legacyImported){
  const legacy:Record<string,string>={};
  // Keep browser values as backup. Inaccessible storage aborts the first import;
  // subsequent launches read SQLite without depending on browser storage.
  for(const key of id?playerKeys:globals){const value=localStorage.getItem(fullKey(id,key));if(value!==null)legacy[key]=value;}
  saved=await backend.LoadPreferences(id,legacy);
 }
 for(const key of id?playerKeys:globals)cache.delete(fullKey(id,key));
 for(const [key,value]of Object.entries(saved.values))cache.set(fullKey(id,key),value);
 ready.add(id);
}
export const preferencesStorage={
 getItem(key:string):string|null {
  if(!host())return localStorage.getItem(key);
  const scope=owner(key);if(!ready.has(scope.id))throw new Error('Preferences have not loaded');
  return cache.get(key)??null;
 },
 setItem(key:string,value:string):Promise<void> {
  const backend=host();if(!backend){try{localStorage.setItem(key,value);return Promise.resolve();}catch(error){return Promise.reject(error);}}
  const scope=owner(key);
  if(!ready.has(scope.id))return Promise.reject(new Error('Preferences have not loaded'));
  const operation=writes.then(async()=>{await backend.SavePreference(scope.id,scope.key,value);cache.set(key,value);});
  writes=operation.catch(()=>{});
  return operation;
 },
};
