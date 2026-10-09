export type FocusDirection = 'left'|'right'|'up'|'down';
type FocusBounds = Pick<DOMRect,'x'|'y'|'width'|'height'>;

// Compare the occupied intervals, not just centers: a full-width setting row
// is directly below every tab even when its center is far to the right.
export function findFocusTarget(direction:FocusDirection,active:FocusBounds,candidates:readonly FocusBounds[]):number {
 const horizontal=direction==='left'||direction==='right';
 const sign=direction==='left'||direction==='up'?-1:1;
 const axes=(rect:FocusBounds)=>horizontal
  ? {start:rect.x,end:rect.x+rect.width,crossStart:rect.y,crossEnd:rect.y+rect.height}
  : {start:rect.y,end:rect.y+rect.height,crossStart:rect.x,crossEnd:rect.x+rect.width};
 const a=axes(active);
 let best=-1,bestBeam=Infinity,bestScore=Infinity;
 candidates.forEach((rect,index)=>{
  const b=axes(rect);
  const forward=((b.start+b.end)-(a.start+a.end))/2*sign;
  if(forward<=4)return;
  const crossGap=Math.max(0,a.crossStart-b.crossEnd,b.crossStart-a.crossEnd);
  const beam=crossGap===0?0:1;
  const forwardGap=Math.max(0,sign===1?b.start-a.end:a.start-b.end);
  const crossCenters=Math.abs((b.crossStart+b.crossEnd-a.crossStart-a.crossEnd)/2);
  const score=forwardGap+crossGap*3+crossCenters*.01;
  if(beam<bestBeam || beam===bestBeam&&score<bestScore){best=index;bestBeam=beam;bestScore=score;}
 });
 return best;
}
export function moveFocus(direction:FocusDirection,root:ParentNode=document) {
 const items=[...root.querySelectorAll<HTMLElement>('[data-nav]')].filter(e=>!e.hasAttribute('disabled')&&!e.closest('[inert]')&&e.getClientRects().length>0);
 const active=document.activeElement as HTMLElement;
 if(!items.includes(active)){items[0]?.focus();return false;}
 const candidates=items.filter(item=>item!==active);
 const index=findFocusTarget(direction,active.getBoundingClientRect(),candidates.map(item=>item.getBoundingClientRect()));
 const best=candidates[index];
 if(!best)return false;
 best.focus({preventScroll:true});best.scrollIntoView({block:'nearest',inline:'nearest'});
 return document.activeElement===best;
}
