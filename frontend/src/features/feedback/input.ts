// Keep this state across React screen/effect changes: a held A/B must not activate twice.
export class GamepadRepeater {
 private held = '';
 private last = 0;
 accept(action: string, now: number) {
  const next = Boolean(action) && (action !== this.held || (now - this.last > 240 && action !== 'confirm' && action !== 'back'));
  this.held = action;
  if (next) this.last = now;
  return next;
 }
 reset() { this.held = ''; this.last = 0; }
}
