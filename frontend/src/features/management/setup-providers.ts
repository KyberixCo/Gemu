import type {CredentialProvider} from '../../library';
export type SetupProvider=CredentialProvider|'manual';
export const setupProviders:Record<SetupProvider,{name:string;purpose:string;docs:string;fields:{name:string;label:string}[];required:string[];canCheck:boolean}>={
 screenscraper:{name:'ScreenScraper',purpose:'Metadatos y arte. Este primer corte consulta juegos de PC Windows.',docs:'https://www.screenscraper.fr/webapi2.php',fields:[{name:'username',label:'Usuario de tu cuenta ScreenScraper'},{name:'password',label:'Contraseña de tu cuenta ScreenScraper'}],required:[],canCheck:true},
 steamgriddb:{name:'SteamGridDB',purpose:'Portadas y fondos adicionales para tus juegos.',docs:'https://www.steamgriddb.com/profile/preferences/api',fields:[{name:'apiKey',label:'Clave API de SteamGridDB'}],required:['apiKey'],canCheck:true},
 steam:{name:'Steam',purpose:'Clave opcional para consultar el perfil y los logros durante la importación.',docs:'https://steamcommunity.com/dev/apikey',fields:[{name:'apiKey',label:'Clave API de Steam'}],required:['apiKey'],canCheck:false},
 igdb:{name:'IGDB',purpose:'Credenciales para metadatos. Su conexión todavía está pendiente de implementación.',docs:'https://api-docs.igdb.com/',fields:[{name:'clientId',label:'Client ID'},{name:'clientSecret',label:'Client secret'}],required:['clientId','clientSecret'],canCheck:false},
 minimax:{name:'MiniMax',purpose:'Clave para el asistente. Las consultas al agente todavía están pendientes.',docs:'',fields:[{name:'apiKey',label:'Clave API'}],required:['apiKey'],canCheck:false},
 openrouter:{name:'OpenRouter',purpose:'Clave para el asistente. Las consultas al agente todavía están pendientes.',docs:'',fields:[{name:'apiKey',label:'Clave API'}],required:['apiKey'],canCheck:false},
 compatible:{name:'API compatible',purpose:'Clave para el asistente. Las consultas al agente todavía están pendientes.',docs:'',fields:[{name:'apiKey',label:'Clave API'}],required:['apiKey'],canCheck:false},
 manual:{name:'Archivos propios',purpose:'Usa tus propios archivos. La publicación de rutas manuales sigue pendiente.',docs:'',fields:[],required:[],canCheck:false},
};
export function providerReadiness(provider:SetupProvider,values:Record<string,string>,saved:readonly string[]){
 const present=(name:string)=>Boolean(values[name])||saved.includes(name);
 const missing=setupProviders[provider].required.filter(name=>!present(name));
 const incompleteAccount=provider==='screenscraper'&&present('username')!==present('password');
 return {ready:missing.length===0&&!incompleteAccount,missing,incompleteAccount};
}
export function providerSteps(provider:SetupProvider,preferences:boolean){
 return ['Introducción',...(provider==='screenscraper'?['Tu cuenta']:setupProviders[provider].fields.length?['Credenciales']:[]),...(setupProviders[provider].canCheck?['Comprobación']:[]),...(preferences?['Preferencias']:[]),'Revisión'];
}
export function sameCredentialValues(first:Record<string,string>,second:Record<string,string>){
 const keys=new Set([...Object.keys(first),...Object.keys(second)]);
 return [...keys].every(key=>(first[key]??'')===(second[key]??''));
}
export function providerActivation(provider:SetupProvider,enabled:boolean,verified:boolean){
 return enabled&&(provider==='manual'||setupProviders[provider].canCheck&&verified);
}
