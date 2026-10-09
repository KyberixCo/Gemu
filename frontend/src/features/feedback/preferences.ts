export type MotionPreference = 'system' | 'reduced' | 'full';
export type FeedbackPreferences = { sound: boolean; volume: number; welcomeSound: boolean; motion: MotionPreference; ambient: boolean; ambientVolume: number; ambientMotion: boolean };
export const defaultPreferences: FeedbackPreferences = { sound: true, volume: 20, welcomeSound: true, motion: 'system', ambient: true, ambientVolume: 25, ambientMotion: true };
export function parsePreferences(raw: string | null): FeedbackPreferences {
 try {
  const value = JSON.parse(raw || 'null');
  return {
   sound: typeof value?.sound === 'boolean' ? value.sound : defaultPreferences.sound,
   volume: typeof value?.volume === 'number' && Number.isFinite(value.volume) ? Math.max(0, Math.min(100, value.volume)) : defaultPreferences.volume,
   welcomeSound: typeof value?.welcomeSound === 'boolean' ? value.welcomeSound : defaultPreferences.welcomeSound,
   ambient: typeof value?.ambient === 'boolean' ? value.ambient : defaultPreferences.ambient,
   ambientVolume: typeof value?.ambientVolume === 'number' && Number.isFinite(value.ambientVolume) ? Math.max(0, Math.min(100, value.ambientVolume)) : defaultPreferences.ambientVolume,
   ambientMotion: typeof value?.ambientMotion === 'boolean' ? value.ambientMotion : defaultPreferences.ambientMotion,
   motion: ['system','reduced','full'].includes(value?.motion) ? value.motion : 'system',
  };
 } catch { return {...defaultPreferences}; }
}
export function isReducedMotion(preference: MotionPreference, systemReduced: boolean) {
 return preference === 'reduced' || (preference === 'system' && systemReduced);
}
