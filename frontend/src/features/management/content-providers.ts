export const contentProviders = {
 igdb: {name:'IGDB',kind:'metadata',description:'Fichas, géneros y datos de juegos.',docs:'https://api-docs.igdb.com/',credentials:['Client ID','Client secret']},
 screenscraper: {name:'ScreenScraper',kind:'metadata',description:'Información y medios para bibliotecas retro.',docs:'https://www.screenscraper.fr/webapi2.php',credentials:['Usuario de tu cuenta ScreenScraper','Contraseña de tu cuenta ScreenScraper']},
 steamgriddb: {name:'SteamGridDB',kind:'art',description:'Portadas, fondos, iconos y logotipos.',docs:'https://www.steamgriddb.com/api/v2',credentials:['Clave API']},
 manual: {name:'Archivos propios',kind:'art',description:'Rutas de arte elegidas en cada ficha. Sin API.',docs:'',credentials:[]},
} as const;
export type ContentProviderId = keyof typeof contentProviders;
export type ContentSettings = Record<ContentProviderId,{enabled:boolean;priority:string}>;
export const defaultContentSettings = ():ContentSettings => Object.fromEntries(Object.keys(contentProviders).map((id,i)=>[id,{enabled:id==='manual',priority:String(i+1)}])) as ContentSettings;
export function readContentSettings(value:unknown):ContentSettings {
 const result=defaultContentSettings();
 if(value && typeof value==='object') for(const id of Object.keys(result) as ContentProviderId[]) {
  const entry=(value as Partial<ContentSettings>)[id];
  if(entry && typeof entry.enabled==='boolean' && /^[1-4]$/.test(entry.priority)) result[id]={enabled:entry.enabled,priority:entry.priority};
 }
 return result;
}
