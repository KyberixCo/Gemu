import {useId,useState} from 'react';
import {artworksForGame,type Game,type GameArtwork} from './catalog';

// Original vector environments, shared by the stage and miniature game tiles.
export function Scene({ game, miniature = false,artwork }: { game: Game; miniature?: boolean;artwork?:GameArtwork }) {
 const instance=useId(),[failed,setFailed]=useState<string>();
 const art=artwork??artworksForGame(game)[0];
 const id = `${miniature ? 'tile' : 'stage'}-${game.id}-${instance.replace(/:/g,'')}`;
 const variant=art.type==='procedural'?art.variant:0;
 const space = game.id === 'astral' || game.id === 'orbit';
 const forest = game.id === 'wild';
 const city = game.id === 'drift';
 const sea = game.id === 'tide';
 if(art.type==='image'&&failed!==art.src)return <img className="scene" src={art.src} style={{objectFit:'cover',objectPosition:art.position??'center'}} alt="" aria-hidden="true" draggable={false} onError={()=>setFailed(art.src)}/>;
 return <svg className="scene" viewBox="0 0 1200 800" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
 <defs>
  <linearGradient id={`${id}-sky`} x2="0.2" y2="1"><stop stopColor={variant===2?'#171d36':space?'#080f1d':city?'#151225':sea?'#244b55':forest?'#7d9b8b':'#453b35'}/><stop offset="1" stopColor={variant===1?'#899ba0':game.color}/></linearGradient>
  <radialGradient id={`${id}-planet`} cx="30%" cy="20%"><stop stopColor="#e5d8a8"/><stop offset=".32" stopColor={game.color}/><stop offset=".72" stopColor="#25333f"/><stop offset="1" stopColor="#080e18"/></radialGradient>
  <linearGradient id={`${id}-ground`} x2="0" y2="1"><stop stopColor={game.color}/><stop offset="1" stopColor="#070b10"/></linearGradient>
 </defs>
 <rect width="1200" height="800" fill={`url(#${id}-sky)`}/>
 {space ? <>
  {Array.from({length:65},(_,i)=><circle key={i} cx={(i*173+23+variant*139)%1200} cy={(i*97+29+variant*71)%640} r={i%4===0?1.7:.8} fill="#d9e5e5" opacity={.3+(i%6)/10}/>)}
  <g transform={`translate(${variant===1?130:variant===2?-90:0} ${variant===1?-85:variant===2?40:0})`}>
  <circle cx="790" cy="284" r={(game.id==='orbit'?160:240)*(variant===1?.75:1)} fill={`url(#${id}-planet)`}/>
  <ellipse cx="790" cy="284" rx={variant===1?280:370} ry="80" fill="none" stroke="#b8c4b0" strokeWidth="2" opacity=".4" transform={`rotate(${variant===2?18:-27} 790 284)`}/>
  </g>
  <path d="M0 610 145 570 220 615 380 535 560 585 780 510 980 568 1200 515V800H0Z" fill="#1c2930"/>
  <path d="M0 712 230 672 450 724 710 610 850 636 1010 600 1200 666V800H0Z" fill="#0b1119"/>
  <path d="M794 612 808 577 817 583 823 614Z" fill="#d6d4bc"/><path d="M807 610v35m12-36 8 34" stroke="#d6d4bc" strokeWidth="5"/>
 </> : city ? <>
  <circle cx={830+variant*65} cy={245-variant*45} r={132-variant*25} fill={variant===2?'#8d9dcc':'#d589ae'} opacity=".6"/>
  {Array.from({length:20},(_,i)=><g key={i}><rect x={i*68-30} y={220+(i*53)%240} width={45+i%3*12} height="390" fill={i%2?'#252438':'#191929'}/>{Array.from({length:7},(_,j)=><path key={j} d={`M${i*68-15} ${290+(i*53)%240+j*24}h18`} stroke={i%3?'#977095':'#77a8b7'} strokeWidth="3"/>)}</g>)}
  <path d="M400 800 740 520h80l360 280" fill="#131522"/><path d="m620 800 150-280m180 280L805 520" stroke="#d692b8" strokeWidth="5"/><path d="m790 560 4 30m-6 20-4 60m-4 25-15 100" stroke="#c1b6b7" strokeWidth="4"/>
  <path d="m700 708 27-53h107l38 53 14 43H690Z" fill="#607f91"/><path d="m735 661-12 35h118l-22-35Z" fill="#152636"/><path d="M702 720h31m101 0h32" stroke="#eddf92" strokeWidth="7"/>
 </> : <>
  <circle cx={820+variant*75} cy={230-variant*40} r={(sea?80:110)-variant*15} fill={variant===2?'#b9c8dd':forest?'#ddd9ab':sea?'#c7d2b0':'#d2a272'} opacity=".8"/>
  <path d={variant===1?'M0 490 180 310 390 450 570 230 750 390 970 210 1200 430V800H0Z':variant===2?'M0 470 160 260 350 410 590 210 800 460 1050 280 1200 400V800H0Z':'M0 530 220 220 330 400 530 160 710 430 890 310 1200 460V800H0Z'} fill={sea?'#34555e':forest?'#49685d':'#615049'}/>
  {variant===0&&<path d="m400 338 130-178 119 174-97-49-28 27-20-35Z" fill="#d5d4bb" opacity=".55"/>}
  <path d="M0 590 190 490 370 550 580 435 850 540 1040 420 1200 465V800H0Z" fill={sea?'#1b343f':forest?'#274a3d':'#332b2b'}/>
  {sea ? <><path d="M0 610Q600 560 1200 600V800H0" fill="#436a70"/>{Array.from({length:12},(_,i)=><path key={i} d={`M${(i*191)%900} ${620+i*15}h${100+i*20}`} stroke="#98b4a9" opacity=".4"/>)}<path d="m820 620 100 4-23 19h-53Z" fill="#131f27"/><path d="M871 618v-98l42 87h-42" stroke="#c9d2bf" fill="#c9d2bf"/></> : <>
  {Array.from({length:23},(_,i)=><path key={i} d={`m${i*62-60} ${580+(i*37)%100} 30 -${120+(i*17)%100} 30 ${120+(i*17)%100}Z`} fill={forest?'#18392e':'#201f22'}/>)}
  <path d="M0 760Q390 580 1200 720V800H0" fill="#0e201b"/>
  {forest?<path d="M700 800Q630 700 800 610L830 610Q710 700 930 800" fill="#a9a484" opacity=".55"/>:<><path d="M790 610V360h55v250m-75-250h90v-25h-90m10-20h65v-25h-65" fill="#171b22"/><path d="m817 607 6-30 8 10 5 23" stroke="#d18d50" strokeWidth="4"/></>}
 </>}
 </>}
 </svg>;
}
