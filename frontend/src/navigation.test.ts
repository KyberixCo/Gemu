import {describe,it,expect} from 'vitest';
import {findFocusTarget,moveMenuFocus} from './navigation';
const rect=(x:number,y:number,width:number,height:number)=>({x,y,width,height});
const tabs=[rect(180,110,100,48),rect(294,110,150,48),rect(458,110,150,48),rect(622,110,120,48)];
const settings=[rect(180,182,950,60),rect(180,260,950,60),rect(180,338,950,60)];
const footer=rect(180,620,130,52);
it('visits menu actions in order, wraps, and recovers missing focus',()=>{
 const focused:number[]=[];
 const items=[0,1,2].map(index=>({focus:()=>focused.push(index),scrollIntoView:()=>{}} as unknown as HTMLElement));
 expect(moveMenuFocus('down',items,items[0])).toBe(true);
 moveMenuFocus('down',items,items[2]);moveMenuFocus('up',items,items[0]);
 moveMenuFocus('right',items,null);moveMenuFocus('left',items,items[1]);
 expect(focused).toEqual([1,0,2,0,0]);
 expect(moveMenuFocus('down',[],null)).toBe(false);
});
describe('spatial focus navigation',()=>{
 it.each(tabs.map((tab,index)=>[index,tab] as const))('moves down from tab %i into the first setting instead of the footer',(_,tab)=>{
  expect(findFocusTarget('down',tab,[...settings,footer])).toBe(0);
 });
 it('keeps left and right navigation within a tab row',()=>{
  expect(findFocusTarget('right',tabs[0],[tabs[1],...settings,footer])).toBe(0);
  expect(findFocusTarget('left',tabs[2],[tabs[1],...settings,footer])).toBe(0);
 });
 it('visits setting rows in order and reaches the footer after the last row',()=>{
  expect(findFocusTarget('down',settings[0],[settings[1],settings[2],footer])).toBe(0);
  expect(findFocusTarget('down',settings[1],[settings[0],settings[2],footer])).toBe(1);
  expect(findFocusTarget('down',settings[2],[...tabs,settings[0],settings[1],footer])).toBe(6);
  expect(findFocusTarget('up',footer,[...settings])).toBe(2);
 });
 it('prefers a candidate in the same column over a closer diagonal candidate',()=>{
  expect(findFocusTarget('down',rect(0,0,100,60),[rect(140,80,100,60),rect(0,180,100,60)])).toBe(1);
 });
 it('uses a diagonal fallback when there is no overlapping target',()=>{
  expect(findFocusTarget('right',rect(0,0,60,60),[rect(100,100,60,60),rect(250,200,60,60)])).toBe(0);
 });
 it('handles controller buttons with partially overlapping bounds',()=>{
  expect(findFocusTarget('down',rect(160,80,42,42),[rect(178,145,42,42),rect(400,120,42,42)])).toBe(0);
 });
 it('ignores targets behind the movement and handles the end of a row',()=>{
  expect(findFocusTarget('right',tabs[3],tabs.slice(0,3))).toBe(-1);
  expect(findFocusTarget('down',footer,tabs)).toBe(-1);
  expect(findFocusTarget('down',footer,[])).toBe(-1);
 });
});
