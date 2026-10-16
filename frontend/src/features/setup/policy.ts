import {preferencesStorage} from '../../persistence';
import {profileKey} from '../../profiles';
import type {Snapshot} from '../../library';
export function setupComplete(){try{return preferencesStorage.getItem(profileKey('gemu.setup.v1'))==='complete';}catch{return false;}}
export function shouldOfferSetup(complete:boolean,snapshot:Pick<Snapshot,'sources'|'games'>){return !complete&&snapshot.sources.length===0&&snapshot.games.length===0;}
