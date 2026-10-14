import {useId} from 'react';
import type {Family} from './controller-profiles';

// Original SVG illustrations, kept in the same coordinate system as the hit targets.
export function ControllerArtwork({family}:{family:Family}){
 const prefix=useId();const body=prefix+'body',metal=prefix+'metal';
 const ps=family==='ps4'||family==='ps5';
 const light=family==='ps5';
 return <svg viewBox="0 0 600 350" aria-hidden="true" className="controller-artwork">
  <defs><linearGradient id={body} x2="0.2" y2="1"><stop stopColor={light?'#fafbff':family==='switch'?'#535c64':'#414952'}/><stop offset="1" stopColor={light?'#aebac8':'#111820'}/></linearGradient><radialGradient id={metal}><stop stopColor="#454e58"/><stop offset=".8" stopColor="#1b222b"/><stop offset="1" stopColor="#090f16"/></radialGradient></defs>
  <ellipse cx="300" cy="317" rx="235" ry="20" fill="#000" opacity=".25"/>
  <path d="M110 35 Q120 9 170 10 L220 17 L215 57 L109 65Z M490 35 Q480 9 430 10 L380 17 L385 57 L491 65Z" fill="#10151c" stroke="#78818b" strokeWidth="2"/>
  {family==='ps5'?<>
   <path d="M151 43 Q109 43 88 80 Q66 135 47 260 Q40 312 72 327 Q101 342 132 294 L185 223 H415 L468 294 Q499 342 528 327 Q560 312 553 260 Q534 135 512 80 Q491 43 449 43Z" fill={`url(#${body})`} stroke="#b9c4d3" strokeWidth="2"/>
   <path d="M178 52 H422 L463 188 L425 242 L480 303 L443 270 L392 213 H208 L157 270 L120 303 L175 242 L137 188Z" fill="#151b25" stroke="#667485"/>
   <path d="M205 51 L195 125 Q300 151 405 125 L395 51" fill="none" stroke="#5b9dff" strokeWidth="4"/>
  </>:family==='ps4'?<>
   <path d="M151 42 Q94 36 77 103 L44 252 Q28 316 68 326 Q105 341 143 277 L181 222 H419 L457 277 Q495 341 532 326 Q572 316 556 252 L523 103 Q506 36 449 42Z" fill={`url(#${body})`} stroke="#71808f" strokeWidth="2"/>
   <path d="M163 212 Q300 243 437 212 L469 292 L420 265 L395 235 H205 L180 265 L131 292Z" fill="#101720" opacity=".8"/>
   <path d="M237 46 H363" stroke="#6ca9ff" strokeWidth="5"/>
  </>:family==='switch'?<>
   <path d="M154 42 Q94 30 77 91 L40 249 Q27 307 67 326 Q104 343 140 288 L184 228 H416 L460 288 Q496 343 533 326 Q573 307 560 249 L523 91 Q506 30 446 42Z" fill={`url(#${body})`} stroke="#84918d" strokeWidth="2"/>
   <path d="M104 83 Q76 144 61 247 Q52 285 80 304 L142 223 L150 87Z M496 83 Q524 144 539 247 Q548 285 520 304 L458 223 L450 87Z" fill="#141e23"/>
   <path d="M220 68 H380 L401 192 Q300 213 199 192Z" fill="#85949a" opacity=".12"/>
   <text x="300" y="154" textAnchor="middle" fill="#a7b7b4" fontSize="11" letterSpacing="3">PRO CONTROLLER</text>
   {[282,294,306,318].map(x=><rect key={x} x={x} y="193" width="5" height="3" rx="1" fill="#b0cd9a"/>)}
  </>:<>
   <path d="M149 43 Q100 30 79 90 L43 240 Q22 303 67 323 Q103 343 144 279 L184 219 H416 L456 279 Q497 343 533 323 Q578 303 557 240 L521 90 Q500 30 451 43Z" fill={`url(#${body})`} stroke="#85909b" strokeWidth="2"/>
   <path d="M152 56 Q300 20 448 56 L422 76 Q300 52 178 76Z" fill="#6b737c" opacity=".35"/>
   <circle cx="300" cy="80" r="15" fill="#b9c2c8"/><path d="M291 70 L309 89 M309 70 L291 89" stroke="#26323b" strokeWidth="4"/>
   <path d="M168 199 Q300 224 432 199" fill="none" stroke="#090f16" strokeWidth="2"/>
  </>}
  {ps&&<><rect x="237" y="51" width="126" height="74" rx={light?16:9} fill={light?'#e0e5ed':'#151c26'} stroke={light?'#929dac':'#53616f'} strokeWidth="2"/>{[0,1,2].map(i=><path key={i} d={`M${285+i*15} 150 h5`} stroke={light?'#69717e':'#8f9aa7'} strokeWidth="2"/>)}<circle cx="300" cy="179" r="9" fill="#17202b" stroke="#8595a5"/></>}
  <circle cx={ps?216:138} cy={ps?238:130} r="32" fill={`url(#${metal})`} stroke="#66717e" strokeWidth="2"/>
  <circle cx="384" cy="238" r="32" fill={`url(#${metal})`} stroke="#66717e" strokeWidth="2"/>
  <path d={ps?'M128 91 H148 V118 H174 V141 H148 V168 H128 V141 H102 V118 H128Z':'M170 199 H190 V226 H216 V249 H190 V276 H170 V249 H144 V226 H170Z'} fill="#101820" stroke="#667481" strokeWidth="2"/>
 </svg>;
}
