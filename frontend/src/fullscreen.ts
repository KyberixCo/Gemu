type DesktopWindow = {
 WindowIsFullscreen:()=>Promise<boolean>;
 WindowFullscreen:()=>void;
 WindowUnfullscreen:()=>void;
};

export async function toggleFullscreen() {
 const desktop=(window as Window&{runtime?:Partial<DesktopWindow>}).runtime;
 if(desktop?.WindowIsFullscreen&&desktop.WindowFullscreen&&desktop.WindowUnfullscreen){
  if(await desktop.WindowIsFullscreen())desktop.WindowUnfullscreen();
  else desktop.WindowFullscreen();
  return;
 }
 if(document.fullscreenElement)await document.exitFullscreen();
 else await document.documentElement.requestFullscreen();
}
