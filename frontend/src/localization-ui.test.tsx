import React from 'react';
import {renderToStaticMarkup} from 'react-dom/server';
import {afterEach,describe,expect,it,vi} from 'vitest';
import {LocalizationProvider} from './localization';
import {TVField,TVScreen} from './features/management/TVControls';
import {Welcome} from './features/feedback/Welcome';
import {WizardFrame} from './features/management/WizardFrame';
import {ContentProviders} from './features/management/ContentProviders';
import {defaultContentSettings} from './features/management/content-providers';
import {ControllerMapper} from './features/management/ControllerMapper';
import {recordInput} from './input-method';

afterEach(()=>{recordInput('keyboard');vi.unstubAllGlobals();});
describe('localized interface',()=>{
 it.each(['en','es'] as const)('renders controls in %s without changing entered values',language=>{
  vi.stubGlobal('localStorage',{getItem:()=>language});
  const markup=renderToStaticMarkup(<LocalizationProvider><TVScreen title="Ajustes" subtitle="Cada emulador conserva su propio perfil."><TVField label="Título" value="Biblioteca" onChange={()=>{}}/><TVField label="Modo de pantalla" value="Pantalla completa" options={['Pantalla completa','Ventana']} onChange={()=>{}}/></TVScreen></LocalizationProvider>);
  expect(markup).toContain(language==='en'?'Settings':'Ajustes');
  expect(markup).toContain(language==='en'?'Fullscreen':'Pantalla completa');
  expect(markup).toContain('Biblioteca');
  if(language==='en')expect(markup).not.toContain('Cada emulador');
 });
 it('translates welcome text and accessible button labels',()=>{
  vi.stubGlobal('localStorage',{getItem:()=> 'en'});
  const markup=renderToStaticMarkup(<LocalizationProvider><Welcome phase="ready" muted ambientMotion={false} active={false} reduced onStart={()=>{}} onSkip={()=>{}} onMute={()=>{}}/></LocalizationProvider>);
  expect(markup).toContain('ALL YOUR WORLDS.');
  expect(markup).toContain('aria-label="Enable sound"');
  expect(markup).toContain('Skip');
 });
 it.each(['es','en'] as const)('renders wizard steps, providers and controller settings in %s',language=>{
  recordInput('gamepad','Xbox Controller',0);
  vi.stubGlobal('localStorage',{getItem:()=>language});
  const markup=renderToStaticMarkup(<LocalizationProvider><WizardFrame title="Preparar Gemu" steps={['Introducción','Credenciales','Revisión']} current={1} actions={null}><ContentProviders kind="art" settings={defaultContentSettings()} onChange={()=>{}} credentials={{}} onCredentials={()=>{}}/><ControllerMapper profiles={{}} onChange={()=>{}} notify={()=>{}} target="Global" onTarget={()=>{}}/></WizardFrame></LocalizationProvider>);
  expect(markup).toContain(language==='en'?'Credentials':'Credenciales');
  expect(markup).toContain(language==='en'?'Custom files':'Archivos propios');
  expect(markup).toContain(language==='en'?'Controller family':'Familia de mando');
  expect(markup).toContain(language==='en'?'Mapping':'Asignación');
  expect(markup).toContain(language==='en'?'aria-label="Setup steps"':'aria-label="Pasos de configuración"');
 });
});
