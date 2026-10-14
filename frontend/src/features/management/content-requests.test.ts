import {expect,it} from 'vitest';
import {defaultContentSettings} from './ContentProviders';
import {contentRequests} from './content-requests';

it('queries only enabled integrations in priority order with their session credentials',()=>{
 const settings=defaultContentSettings();
 expect(contentRequests(settings,{'steamgriddb:0':'session-key'},'es')).toEqual([]);
 settings.igdb.enabled=true;settings.steamgriddb={enabled:true,priority:'1'};settings.screenscraper={enabled:true,priority:'2'};
 const credentials={'steamgriddb:0':'session-key','screenscraper:0':'user','screenscraper:1':'password','igdb:0':'other-key'};
 expect(contentRequests(settings,credentials,'en')).toEqual([
  {id:'steamgriddb',priority:1,language:'en',credentials:{apiKey:'session-key'}},
  {id:'screenscraper',priority:2,language:'en',credentials:{username:'user',password:'password'}},
 ]);
 expect(JSON.stringify(settings)).not.toContain('session-key');
});
