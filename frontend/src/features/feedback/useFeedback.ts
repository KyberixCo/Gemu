import {profileKey} from '../../profiles';
import { useEffect, useState } from 'react';
import { interfaceAudio } from './audio';
import { isReducedMotion, parsePreferences, type FeedbackPreferences } from './preferences';
const KEY = 'gemu.feedback';
export function useFeedback(onError: (message:string) => void) {
 const [preferences,setPreferences] = useState<FeedbackPreferences>(() => {
  try { return parsePreferences(localStorage.getItem(profileKey(KEY))); } catch { return parsePreferences(null); }
 });
 const [active,setActive] = useState(() => !document.hidden && document.hasFocus());
 const [systemReduced,setSystemReduced] = useState(() => matchMedia('(prefers-reduced-motion: reduce)').matches);
 useEffect(() => {
  const media = matchMedia('(prefers-reduced-motion: reduce)');
  const update = () => setSystemReduced(media.matches);
  media.addEventListener('change',update);return () => media.removeEventListener('change',update);
 },[]);
 useEffect(() => { interfaceAudio.configure(preferences); },[preferences]);
 useEffect(() => {
  void interfaceAudio.preload();
  const sync = () => {const foreground=!document.hidden && document.hasFocus();setActive(foreground);interfaceAudio.setActive(foreground);};
  const unlock = () => { void interfaceAudio.unlock(); };
  sync();
  window.addEventListener('focus',sync);window.addEventListener('blur',sync);document.addEventListener('visibilitychange',sync);
  window.addEventListener('pointerdown',unlock);window.addEventListener('keydown',unlock);
  return () => {
   interfaceAudio.stop();window.removeEventListener('focus',sync);window.removeEventListener('blur',sync);document.removeEventListener('visibilitychange',sync);
   window.removeEventListener('pointerdown',unlock);window.removeEventListener('keydown',unlock);
  };
 },[]);
 const updatePreferences = (patch: Partial<FeedbackPreferences>) => {
  const next = {...preferences,...patch};
  interfaceAudio.configure(next);setPreferences(next);
  try { localStorage.setItem(profileKey(KEY),JSON.stringify(next)); } catch { onError('No se pudieron guardar tus preferencias.'); }
 };
 return {preferences,updatePreferences,active,reduced:isReducedMotion(preferences.motion,systemReduced)};
}
