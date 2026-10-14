import type {Family,Control} from './controller-profiles';

// Match hardware IDs before names: Bluetooth often reports only "Wireless Controller".
export function detectControllerFamily(id:string):Family|null{
 const name=id.toLowerCase();
 const chrome=name.match(/vendor:\s*([0-9a-f]{4}).*product:\s*([0-9a-f]{4})/);
 const firefox=name.match(/^([0-9a-f]{4})-([0-9a-f]{4})-/);
 const pair=chrome??firefox;
 if(pair){
  const [,vendor,product]=pair;
  if(vendor==='054c'){
   if(['0ce6','0df2'].includes(product))return 'ps5';
   if(['05c4','09cc','0ba0'].includes(product))return 'ps4';
  }
  if(vendor==='057e'&&['2009','200e'].includes(product))return 'switch';
  if(vendor==='045e'&&['028e','028f','0291','02a1','02d1','02dd','02e0','02e3','02ea','02fd','0719','0b00','0b05','0b06','0b0a','0b0c','0b12','0b13','0b20','0b21','0b22'].includes(product))return 'xbox';
 }
 if(/dualsense|\bps5\b|playstation\s*5/.test(name))return 'ps5';
 if(/dualshock\s*4|\bps4\b|playstation\s*4/.test(name))return 'ps4';
 if(/switch.*pro|pro.*controller.*switch|joy-?con.*(?:pair|combined|l\+r)/.test(name))return 'switch';
 if(/xbox|xinput|x-input/.test(name))return 'xbox';
 return null;
}

export function controlPosition(family:Family,id:Control):{x:number;y:number}{
 const ps=family==='ps4'||family==='ps5';
 const positions:Record<Control,[number,number]>={
  south:[79,49],east:[86,37],west:[72,37],north:[79,25],
  lb:[23,13],rb:[77,13],lt:[23,4],rt:[77,4],
  up:[ps?23:30,ps?27:58],down:[ps?23:30,ps?47:78],
  left:[ps?17:24,ps?37:68],right:[ps?29:36,ps?37:68],
  ls:[ps?36:23,ps?68:37],rs:[64,68],select:[ps?35:43,ps?23:32],start:[ps?65:57,ps?23:32],
 };
 const [x,y]=positions[id];return {x,y};
}
