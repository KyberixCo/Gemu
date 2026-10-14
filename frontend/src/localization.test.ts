import {describe,expect,it} from 'vitest';
import {resolveLanguage,translate} from './localization';

describe('localization',()=>{
 it('uses the first supported browser preference and respects saved choices',()=>{
  expect(resolveLanguage(null,['es-CO','en-US'])).toBe('es');
  expect(resolveLanguage(null,['fr','en-GB','es'])).toBe('en');
  expect(resolveLanguage('es',['en-US'])).toBe('es');
  expect(resolveLanguage('invalid',['en-US'])).toBe('en');
  expect(resolveLanguage(null,[])).toBe('es');
 });
 it('interpolates values literally in either language',()=>{
  const message='Guardar arte de {count} juegos';
  expect(translate(message,'en',{count:2})).toBe('Save artwork for 2 games');
  expect(translate(message,'es',{count:0})).toBe('Guardar arte de 0 juegos');
  expect(translate('Configurar {name}','en',{name:'$& {count}'})).toBe('Set up $& {count}');
  expect(translate(message,'en')).toBe('Save artwork for {count} games');
 });
 it('translates Spanish interface text to English',()=>{
  expect(translate('Completar arte faltante','en')).toBe('Fill in missing artwork');
  expect(translate('Añadir a favoritos','en')).toBe('Add to favorites');
  expect(translate('Elige el idioma de Gemu','en')).toBe('Choose Gemu’s language');
  expect(translate('Steam aporta el arte básico. Configura SteamGridDB para ampliarlo o ScreenScraper para obtener metadatos y arte.','en')).toContain('Steam provides basic artwork');
 });
 it('keeps Spanish text in Spanish and leaves unknown strings intact',()=>{
  expect(translate('Completar arte faltante','es')).toBe('Completar arte faltante');
  expect(translate('A game title','en')).toBe('A game title');
  for(const text of ['constructor','toString','__proto__'])expect(translate(text,'en')).toBe(text);
 });
});
