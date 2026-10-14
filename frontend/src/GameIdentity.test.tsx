import {expect,it} from 'vitest';
import {renderToStaticMarkup} from 'react-dom/server';
import {games} from './catalog';
import {GameTitle,GameIcon} from './GameIdentity';

it('uses the selected logo as an accessible visual title without duplicating text',()=>{
 const game={...games[0],logo:'/library-assets/logo'};
 const html=renderToStaticMarkup(<GameTitle game={game}/>);
 expect(html).toContain('<h1');expect(html).toContain('alt="'+game.title+'"');expect(html).not.toContain('<span>');
});
it('uses the localized text title instead of an Arabic logo for Mouse: P.I. for Hire',()=>{
 const game={...games[0],title:'MOUSE: P.I. FOR HIRE',logo:'/library-assets/arabic-logo'};
 const html=renderToStaticMarkup(<GameTitle game={game}/>);
 expect(html).toContain('<span>MOUSE: </span>');
 expect(html).toContain('<span>P.I. </span>');
 expect(html).not.toContain('arabic-logo');
});
it('shows a manually selected logo for Mouse: P.I. for Hire',()=>{
 const game={...games[0],title:'MOUSE: P.I. FOR HIRE',logo:'/library-assets/selected-logo',logoManuallySelected:true};
 const html=renderToStaticMarkup(<GameTitle game={game}/>);
 expect(html).toContain('src="/library-assets/selected-logo"');
});
it('keeps a readable title and avoids broken placeholders when no assets exist',()=>{
 expect(renderToStaticMarkup(<GameTitle game={games[0]} level={2}/>)).toContain('<h2');
 expect(renderToStaticMarkup(<GameTitle game={games[0]}/>)).not.toContain('<img');
 expect(renderToStaticMarkup(<GameIcon game={games[0]}/>)).toBe('');
});
it('uses icons as decorative identity alongside the readable game name',()=>{
 const html=renderToStaticMarkup(<GameIcon game={{...games[0],icon:'/library-assets/icon'}}/>);
 expect(html).toContain('src="/library-assets/icon"');expect(html).toContain('aria-hidden="true"');expect(html).toContain('alt=""');
});
