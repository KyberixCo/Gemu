import { useEffect, useRef } from 'react';
import { AmbientField } from './AmbientField';
import { Volume2, VolumeX } from 'lucide-react';
export type WelcomePhase = 'ready' | 'playing' | 'done';
export function Welcome({ phase, muted, onStart, onSkip, onMute, ambientMotion, active, reduced }: { phase: WelcomePhase; ambientMotion: boolean; active: boolean; reduced: boolean; muted: boolean; onStart: () => void; onSkip: () => void; onMute: () => void }) {
 const start = useRef<HTMLButtonElement>(null);
 useEffect(() => { if (phase === 'ready') start.current?.focus(); }, [phase]);
 if (phase === 'done') return null;
 return <section className={`welcome welcome-${phase}`} aria-label="Bienvenida a Gemu">
  <AmbientField enabled={ambientMotion} active={active} reduced={reduced} welcome/>
  <div className="welcome-grid" aria-hidden="true"/>
  <span className="welcome-system">KYBERIX / GEMU</span>
  <div className="welcome-core">
   <div className="welcome-emblem" aria-hidden="true"><i/><i/><i/><i/><span className="emblem-center"/></div>
   <h1 className="welcome-word" aria-label="Gemu">{'gemu'.split('').map((letter,i) => <span key={i} style={{'--letter':i} as React.CSSProperties}>{letter}</span>)}<b>.</b></h1>
   <div className="welcome-line" aria-hidden="true"/>
   <p className="welcome-caption">TODOS TUS MUNDOS.</p>
  </div>
  <div className="welcome-controls">
   {phase === 'ready' ? <button data-nav ref={start} className="welcome-start" onClick={onStart}><span className="pad-key">A</span> Entrar <kbd>ENTER</kbd></button> : <span className="welcome-loading" role="status">Entrando a tu biblioteca</span>}
   <div className="welcome-options"><button data-nav className="welcome-mute" aria-label={muted?'Activar sonido':'Silenciar sonido'} aria-pressed={muted} onClick={onMute}>{muted?<VolumeX size={18}/>:<Volume2 size={18}/>}<span>{muted?'Sin sonido':'Con sonido'}</span></button><button data-nav onClick={onSkip}><span className="pad-key">B</span> Omitir</button></div>
  </div>
 </section>;
}
