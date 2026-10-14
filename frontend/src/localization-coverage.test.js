import {readFileSync,readdirSync} from 'node:fs';
import {resolve} from 'node:path';
import {expect,it} from 'vitest';
import {translate} from './localization';

// Catch new untranslated host errors before they reach a wizard or a saved
// library diagnostic. Variable technical details are excluded from the key.
it('covers Spanish errors and player messages emitted by the host',()=>{
 const root=resolve(import.meta.dirname,'../..');
 const files=['app.go',...readdirSync(resolve(root,'internal'),{recursive:true}).filter(file=>typeof file==='string'&&file.endsWith('.go')&&!file.endsWith('_test.go')).map(file=>'internal/'+file)];
 const missing=new Set();
 for(const file of files){
  const source=readFileSync(resolve(root,file),'utf8');
  const regex=/(?:fmt\.Errorf\(|errors\.New\(|(?:\.Message|message)\s*=\s*)"([^"\n]+)"/g;
  for(const match of source.matchAll(regex)){
   let text=match[1].replace(/: %[wsq].*$/,'').replace('%d','503').trim();
   if(!/[áéíóúñ]|\b(selecciona|no|sin|falta|completa|servicio|fuente|hay|usa|elige|cierra|termina|proveedor|introduce)\b/i.test(text))continue;
   if(translate(text,'en')===text)missing.add(text);
  }
 }
 expect([...missing]).toEqual([]);
});
