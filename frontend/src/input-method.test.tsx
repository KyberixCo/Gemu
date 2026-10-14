import React from 'react';
import {renderToStaticMarkup} from 'react-dom/server';
import {afterEach,expect,it,vi} from 'vitest';
import {inputMethod,recordInput,reconcileInput,gamepadHint} from './input-method';
import {ActionHint} from './ActionHint';
import {LocalizationProvider} from './localization';
import {ControllerMapper} from './features/management/ControllerMapper';
afterEach(()=>{recordInput('keyboard');vi.unstubAllGlobals();});
it('keeps keyboard/mouse active with an idle connected controller',()=>{
 recordInput('gamepad','Xbox Controller',1);expect(inputMethod().family).toBe('xbox');
 recordInput('mouse');reconcileInput([{index:1,connected:true}]);expect(inputMethod().kind).toBe('mouse');
 recordInput('keyboard');expect(inputMethod().kind).toBe('keyboard');
});
it('falls back when the active controller disconnects and labels physical buttons',()=>{
 recordInput('gamepad','DualSense',2);expect(gamepadHint(inputMethod().family,'confirm')).toBe('×');
 reconcileInput([{index:1,connected:true}]);expect(inputMethod().kind).toBe('keyboard');
 expect(gamepadHint('switch','confirm')).toBe('B');expect(gamepadHint('switch','back')).toBe('A');
 expect(gamepadHint('xbox','confirm')).toBe('A');expect(gamepadHint(null,'confirm')).toBe('●');
});
it('renders desktop hints and keyboard mappings without Xbox prompts',()=>{
 vi.stubGlobal('localStorage',{getItem:()=> 'es'});recordInput('keyboard');
 const markup=renderToStaticMarkup(<LocalizationProvider><ActionHint action="confirm"/><ActionHint action="back"/><ControllerMapper profiles={{}} onChange={()=>{}} notify={()=>{}} target="Global" onTarget={()=>{}}/></LocalizationProvider>);
 expect(markup).toContain('<kbd>Enter</kbd>');expect(markup).toContain('<kbd>Esc</kbd>');expect(markup).toContain('Navegación de Gemu');expect(markup).not.toContain('controller-diagram');expect(markup).not.toContain('Familia de mando');expect(markup).not.toContain('pad-key');
 recordInput('mouse');expect(renderToStaticMarkup(<LocalizationProvider><ActionHint action="confirm"/></LocalizationProvider>)).toContain('<kbd>Clic</kbd>');
 recordInput('gamepad','DualSense',0);expect(renderToStaticMarkup(<LocalizationProvider><ActionHint action="confirm"/></LocalizationProvider>)).toContain('×');
});
