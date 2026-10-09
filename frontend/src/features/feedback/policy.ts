export type SoundName = 'welcome' | 'navigate' | 'section' | 'open' | 'close' | 'favorite-on' | 'favorite-off' | 'confirm' | 'error';
export const soundNames: SoundName[] = ['welcome','navigate','section','open','close','favorite-on','favorite-off','confirm','error'];
export const priority = (name: SoundName) => name === 'error' ? 3 : name === 'welcome' ? 2 : name === 'navigate' ? 0 : 1;
// Pure admission policy: no backlog and no duplicate ticks when a direction hits a boundary.
export class SoundPolicy {
 private lastNavigation = -Infinity;
 accept(name: SoundName, now: number, enabled: boolean, changed = true) {
  if (!enabled || !changed) return false;
  if (name === 'navigate') {
   if (now - this.lastNavigation < 90) return false;
   this.lastNavigation = now;
  }
  return true;
 }
 reset() { this.lastNavigation = -Infinity; }
}
