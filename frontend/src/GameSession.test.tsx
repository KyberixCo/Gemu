import React from 'react';
import {renderToStaticMarkup} from 'react-dom/server';
import {afterEach,expect,it,vi} from 'vitest';
import {SessionMenu,type Session} from './GameSession';
import {LocalizationProvider} from './localization';
import {recordInput} from './input-method';
import {games} from './catalog';

const session:Session={gameId:games[0].id,title:'My current game',state:'running',menu:true,background:false,message:''};
function menu(overrides:{state?:string;confirm?:boolean;busy?:boolean;error?:string}={}){
 return renderToStaticMarkup(<LocalizationProvider><SessionMenu session={{...session,state:overrides.state??'running'}} game={{...games[0],title:'Different library title'}} confirm={overrides.confirm??false} busy={overrides.busy??false} error={overrides.error??''} onResume={()=>{}} onHome={()=>{}} onClose={()=>{}} onCancel={()=>{}} onConfirm={()=>{}}/></LocalizationProvider>);
}
const buttons=(markup:string)=>[...markup.matchAll(/<button\b[^>]*>[\s\S]*?<\/button>/g)].map(match=>match[0]);
afterEach(()=>{recordInput('keyboard');vi.unstubAllGlobals();});
it.each(['es','en'])('renders the running game and localized controls in %s',language=>{
 vi.stubGlobal('localStorage',{getItem:()=>language});
 const markup=menu();
 expect(markup).toContain('role="dialog"');
 expect(markup).toContain('aria-modal="true"');
 expect(markup).toContain('My current game');
 expect(markup).not.toContain('Different library title');
 expect(buttons(markup)).toHaveLength(3);
 expect(buttons(markup)[0]).toContain(language==='es'?'Continuar jugando':'Resume game');
 expect(buttons(markup)[1]).toContain(language==='es'?'Volver a Gemu':'Return to Gemu');
 expect(markup).toContain('<kbd>↑ ↓</kbd>');
});
it('offers the library while launch is unconfirmed and disables unsafe actions',()=>{
 const actions=buttons(menu({state:'requested'}));
 expect(actions[0]).toContain('disabled=""');
 expect(actions[1]).not.toContain('disabled=""');
 expect(actions[2]).toContain('disabled=""');
});
it('allows resuming a closing game but prevents another close request',()=>{
 const actions=buttons(menu({state:'closing'}));
 expect(actions[0]).not.toContain('disabled=""');
 expect(actions[2]).toContain('disabled=""');
});
it('makes cancellation the first action in the destructive confirmation',()=>{
 vi.stubGlobal('localStorage',{getItem:()=> 'en'});
 const markup=menu({confirm:true});
 expect(markup).toContain('role="alertdialog"');
 expect(markup).toContain('aria-describedby="session-warning"');
 expect(markup).toContain('Unsaved progress may be lost.');
 expect(buttons(markup)).toHaveLength(2);
 expect(buttons(markup)[0]).toContain('data-session-action="cancel"');
 expect(buttons(markup)[0]).toContain('Cancel');
 expect(buttons(markup)[1]).toContain('Confirm close');
});
it('disables every action during a host request and exposes errors in the menu',()=>{
 const markup=menu({busy:true,error:'Host failed'});
 expect(buttons(markup).every(button=>button.includes('disabled=""'))).toBe(true);
 expect(markup).toContain('aria-busy="true"');
 expect(markup).toContain('role="alert">Host failed');
});
it('shows physical controller hints when navigating with a gamepad',()=>{
 recordInput('gamepad','Xbox Controller',0);
 const markup=menu();
 expect(markup).toContain('data-input="gamepad"');
 expect(markup).toContain('>A</span>');
 expect(markup).toContain('>B</span>');
 expect(markup).toContain('Xbox / View + Menu');
});
