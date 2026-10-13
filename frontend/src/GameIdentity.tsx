import {useState} from 'react';
import type {Game} from './catalog';

// A logo is the visual title; the textual title remains the accessible fallback.
export function GameTitle({game,level=1,className=''}:{game:Game;level?:1|2;className?:string}) {
 const [failed,setFailed]=useState<string>();
 const Heading=level===1?'h1':'h2';
 return <Heading className={className}>{game.logo&&failed!==game.logo?
  <img className="game-logo" src={game.logo} alt={game.title} onError={()=>setFailed(game.logo)}/>:
  game.title.split(' ').map((word,index)=><span key={index}>{word} </span>)}</Heading>;
}
export function GameIcon({game}:{game:Game}) {
 const [failed,setFailed]=useState<string>();
 return game.icon&&failed!==game.icon?<img className="game-icon" src={game.icon} alt="" aria-hidden="true" onError={()=>setFailed(game.icon)}/>:null;
}
