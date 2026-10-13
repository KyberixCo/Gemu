export type GemuProfile={id:string;name:string};
export type ProfileHost={GetProfiles:()=>Promise<GemuProfile[]>;CreateProfile:(name:string)=>Promise<GemuProfile>;SelectProfile:(id:string)=>Promise<void>};
let profileID='';
export function setStorageProfile(id:string){profileID=id;}
export function profileKey(key:string){if(!profileID)throw new Error('Selecciona un perfil de Gemu.');return `gemu.profile.${profileID}.${key}`;}
export function startupAction(profiles:GemuProfile[]):'create'|'select'|'enter'{return profiles.length===0?'create':profiles.length===1?'enter':'select';}
