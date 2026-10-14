import {useLocalization} from '../../localization';
import { useEffect, useRef } from 'react';
import { AmbientField } from './AmbientField';
import { Volume2, VolumeX } from 'lucide-react';
import {ActionHint} from '../../ActionHint';
export type WelcomePhase = 'ready' | 'playing' | 'done';
export function Welcome({ phase, muted, onStart, onSkip, onMute, ambientMotion, active, reduced }: { phase: WelcomePhase; ambientMotion: boolean; active: boolean; reduced: boolean; muted: boolean; onStart: () => void; onSkip: () => void; onMute: () => void }) {
 const {t}=useLocalization();
 const start = useRef<HTMLButtonElement>(null);
 useEffect(() => { if (phase === 'ready') start.current?.focus(); }, [phase]);
 if (phase === 'done') return null;
 return <section className={`welcome welcome-${phase}`} aria-label={t("Bienvenida a Gemu")}>
  <AmbientField enabled={ambientMotion} active={active} reduced={reduced} welcome/>
  <div className="welcome-grid" aria-hidden="true"/>
  <span className="welcome-system">KYBERIX / GEMU</span>
  <div className="welcome-core">
   <div className="welcome-emblem" aria-hidden="true"><i/><i/><i/><i/><span className="emblem-center"/></div>
   <h1 className="welcome-word" aria-label={"Gemu"}>{'gemu'.split('').map((letter,i) => <span key={i} style={{'--letter':i} as React.CSSProperties}>{letter}</span>)}<b>.</b></h1>
   <div className="welcome-line" aria-hidden="true"/>
   <p className="welcome-caption">{t("TODOS TUS MUNDOS.")}</p>
  </div>
  <div className="welcome-controls">
   {phase === 'ready' ? <button data-nav ref={start} className="welcome-start" onClick={onStart}><ActionHint action="confirm"/>{t("Entrar")}</button> : <span className="welcome-loading" role="status">{t("Entrando a tu biblioteca")}</span>}
   <div className="welcome-options"><button data-nav className="welcome-mute" aria-label={muted?t("Activar sonido"):t("Silenciar sonido")} aria-pressed={muted} onClick={onMute}>{muted?<VolumeX size={18}/>:<Volume2 size={18}/>}<span>{muted?t("Sin sonido"):t("Con sonido")}</span></button><button data-nav onClick={onSkip}><ActionHint action="back"/>{t("Omitir")}</button></div>
  </div>
 </section>;
}
