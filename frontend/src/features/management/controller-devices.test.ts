import {describe,it,expect} from 'vitest';
import {detectControllerFamily,controlPosition} from './controller-devices';
import {controls} from './controller-profiles';
import {renderToStaticMarkup} from 'react-dom/server';
import {createElement} from 'react';
import {ControllerArtwork} from './ControllerArtwork';
describe('commercial controller detection',()=>{
 it.each([
  ['Xbox 360 Controller (XInput STANDARD GAMEPAD)','xbox'],
  ['Xbox Wireless Controller','xbox'],
  ['Xbox Series X Controller','xbox'],
  ['Controller (STANDARD GAMEPAD Vendor: 045e Product: 0b13)','xbox'],
  ['045e-028e-Controller','xbox'],
  ['Wireless Controller (STANDARD GAMEPAD Vendor: 054c Product: 05c4)','ps4'],
  ['Wireless Controller (STANDARD GAMEPAD Vendor: 054C Product: 09CC)','ps4'],
  ['054c-0ba0-Wireless Controller','ps4'],
  ['Wireless Controller (STANDARD GAMEPAD Vendor: 054c Product: 0ce6)','ps5'],
  ['054c-0df2-Wireless Controller','ps5'],
  ['DualSense Wireless Controller','ps5'],
  ['DUALSHOCK 4 Wireless Controller','ps4'],
  ['Pro Controller (STANDARD GAMEPAD Vendor: 057e Product: 2009)','switch'],
  ['Nintendo Switch Pro Controller','switch'],
  ['Joy-Con (L+R)','switch'],
 ])('identifies %s',(id,family)=>expect(detectControllerFamily(id)).toBe(family));
 it.each(['Wireless Controller','Generic USB Joystick','054c-0268-PLAYSTATION 3 Controller','Joy-Con (L)','Joy-Con (R)','045e-0000-Unknown device'])('does not guess ambiguous or unsupported devices: %s',id=>{
  expect(detectControllerFamily(id)).toBeNull();
 });
 it('gives authoritative hardware IDs priority over ambiguous names',()=>{
  expect(detectControllerFamily('PS4 Wireless Controller (Vendor: 054c Product: 0ce6)')).toBe('ps5');
 });
});
describe('controller illustrations',()=>{
 it('renders distinct artwork for each family and keeps overlays within the canvas',()=>{
  const artwork=new Set<string>();
  for(const family of ['xbox','ps4','ps5','switch'] as const){
   artwork.add(renderToStaticMarkup(createElement(ControllerArtwork,{family})));
   for(const control of controls){const {x,y}=controlPosition(family,control.id);expect(x).toBeGreaterThan(0);expect(x).toBeLessThan(100);expect(y).toBeGreaterThan(0);expect(y).toBeLessThan(100);}
  }
  expect(artwork.size).toBe(4);
  expect(controlPosition('xbox','ls').y).toBeLessThan(controlPosition('ps5','ls').y);
  expect(controlPosition('ps4','ls')).toEqual(controlPosition('ps5','ls'));
 });
});
