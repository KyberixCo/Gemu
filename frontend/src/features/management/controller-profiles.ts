export const families = { xbox:'Xbox', switch:'Nintendo Switch · Pro / Joy-Con combinados', ps4:'PlayStation 4 · DualShock 4', ps5:'PlayStation 5 · DualSense' };
export type Family = keyof typeof families;
export const controls = [
 {id:'south',x:78,y:46},{id:'east',x:85,y:35},{id:'west',x:71,y:35},{id:'north',x:78,y:24},
 {id:'lb',x:22,y:12},{id:'rb',x:78,y:12},{id:'lt',x:22,y:3},{id:'rt',x:78,y:3},
 {id:'up',x:27,y:57},{id:'down',x:27,y:77},{id:'left',x:20,y:67},{id:'right',x:34,y:67},
 {id:'ls',x:27,y:35},{id:'rs',x:64,y:67},{id:'select',x:43,y:33},{id:'start',x:57,y:33},
] as const;
export type Control = typeof controls[number]['id'];
export const actions = ['Sin asignar','A / Cruz','B / Círculo','X / Cuadrado','Y / Triángulo','Arriba','Abajo','Izquierda','Derecha','L1 / L','R1 / R','L2 / ZL','R2 / ZR','L3','R3','Select','Start','Menú del emulador','Guardar estado','Cargar estado','Rebobinar','Avance rápido'];
export type ControllerProfile = { bindings:Record<Control,string>;deadzone:number;sensitivity:number;invertY:boolean };
export type ControllerProfiles = Record<string,ControllerProfile>;
export function label(family:Family,id:Control):string {
 const faces = family==='switch'?['B','A','Y','X']:family==='xbox'?['A','B','X','Y']:['×','○','□','△'];
 const names:Record<Control,string> = {south:faces[0],east:faces[1],west:faces[2],north:faces[3],lb:family==='switch'?'L':family==='xbox'?'LB':'L1',rb:family==='switch'?'R':family==='xbox'?'RB':'R1',lt:family==='switch'?'ZL':family==='xbox'?'LT':'L2',rt:family==='switch'?'ZR':family==='xbox'?'RT':'R2',up:'↑',down:'↓',left:'←',right:'→',ls:'L3',rs:'R3',select:family==='switch'?'−':family==='xbox'?'Vista':family==='ps4'?'Share':'Crear',start:family==='switch'?'+':family==='xbox'?'Menú':'Options'};
 return names[id];
}
export function defaultProfile():ControllerProfile {return {bindings:{south:'A / Cruz',east:'B / Círculo',west:'X / Cuadrado',north:'Y / Triángulo',up:'Arriba',down:'Abajo',left:'Izquierda',right:'Derecha',lb:'L1 / L',rb:'R1 / R',lt:'L2 / ZL',rt:'R2 / ZR',ls:'L3',rs:'R3',select:'Select',start:'Start'},deadzone:15,sensitivity:100,invertY:false};}
export function validProfiles(value:unknown):value is ControllerProfiles {
 if(!value||typeof value!=='object'||Array.isArray(value))return false;
 return Object.entries(value).every(([key,p])=>{
  const profile=p as ControllerProfile;
  return /^(Global|RetroArch|Dolphin|PCSX2|PPSSPP):(xbox|switch|ps4|ps5)(:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})?$/i.test(key)&&profile&&Number.isFinite(profile.deadzone)&&profile.deadzone>=0&&profile.deadzone<=50&&Number.isFinite(profile.sensitivity)&&profile.sensitivity>=50&&profile.sensitivity<=150&&typeof profile.invertY==='boolean'&&controls.every(c=>actions.includes(profile.bindings?.[c.id]));
 });
}
