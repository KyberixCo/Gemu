import {describe,expect,it} from 'vitest';
import {translateHostMessage} from './host-messages';
import {translate} from './localization';

describe('host diagnostics',()=>{
 it('translates nested errors while retaining game titles and technical details',()=>{
  const raw='Mouse: P.I. For Hire: SteamGridDB · credenciales rechazadas o acceso no permitido: proveedor no disponible';
  expect(translateHostMessage(raw,'en')).toBe('Mouse: P.I. For Hire: SteamGridDB · credentials rejected or access denied: provider unavailable');
  expect(translateHostMessage(raw,'es')).toBe(raw);
  expect(translateHostMessage('Error: biblioteca guardada inválida: invalid JSON at byte 4','en')).toBe('Error: invalid saved library: invalid JSON at byte 4');
 });
 it('keeps paths, unknown provider responses and inserted characters intact',()=>{
  expect(translateHostMessage('la carpeta no contiene una biblioteca Steam accesible (steamapps): C:\\Games\\Steam','en')).toBe('the folder does not contain an accessible Steam library (steamapps): C:\\Games\\Steam');
  expect(translateHostMessage('No se pudo descargar una imagen de $&','en')).toBe('Could not download an image from $&');
  expect(translateHostMessage('unknown network response','en')).toBe('unknown network response');
 });
 it('handles HTTP errors and cached achievement messages',()=>{
  expect(translateHostMessage('arte HTTP 404','en')).toBe('artwork HTTP 404');
  expect(translateHostMessage('servicio HTTP 503','en')).toBe('service HTTP 503');
  expect(translateHostMessage('Se muestran los últimos logros guardados. Logros sincronizados desde Steam.','en')).toBe('Showing the last saved achievements. Achievements synced from Steam.');
 });
 it('translates stored errors again for the current language',()=>{
  const raw='ya existe un perfil con ese nombre';
  expect(translate(raw,'en')).toBe('a profile with that name already exists');
  expect(translate(raw,'es')).toBe(raw);
 });
 it('translates known technical errors into Spanish and preserves unknown OS details',()=>{
  expect(translate('C:\\Steam\\manifest.acf: invalid installation flags','es')).toBe('C:\\Steam\\manifest.acf: Indicadores de instalación inválidos');
  expect(translate('missing value for "appid"','es')).toBe('Falta un valor para "appid"');
  expect(translate('access denied at C:\\Games','es')).toBe('access denied at C:\\Games');
 });
});
