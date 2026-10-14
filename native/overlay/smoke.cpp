#include "protocol.h"
#include <d3d11.h>
#include <dxgi.h>
#include <cstdio>
#include <cwchar>
#include <cmath>

static bool running=true;
static LRESULT CALLBACK wndproc(HWND hwnd,UINT message,WPARAM wp,LPARAM lp){
 if(message==WM_CLOSE){running=false;return 0;}
 return DefWindowProcW(hwnd,message,wp,lp);
}
int main(int argc,char** argv){
 bool exclusive=argc>1&&strcmp(argv[1],"--exclusive")==0;
 const char* report=argc>2?argv[2]:"overlay-report.json";
 auto instance=GetModuleHandleW(nullptr);WNDCLASSW cls{};cls.hInstance=instance;cls.lpfnWndProc=wndproc;cls.lpszClassName=L"GemuOverlaySmoke";RegisterClassW(&cls);
 DEVMODEW initial{};initial.dmSize=sizeof(initial);EnumDisplaySettingsW(nullptr,ENUM_CURRENT_SETTINGS,&initial);
 HWND window=CreateWindowExW(0,cls.lpszClassName,L"Gemu · DirectX 11 overlay validation",WS_OVERLAPPEDWINDOW,60,60,1280,720,nullptr,nullptr,instance,nullptr);
 DXGI_SWAP_CHAIN_DESC desc{};desc.BufferCount=1;desc.BufferDesc.Width=exclusive?initial.dmPelsWidth:1280;desc.BufferDesc.Height=exclusive?initial.dmPelsHeight:720;desc.BufferDesc.Format=DXGI_FORMAT_R8G8B8A8_UNORM;desc.BufferUsage=DXGI_USAGE_RENDER_TARGET_OUTPUT;desc.OutputWindow=window;desc.SampleDesc.Count=1;desc.Windowed=TRUE;desc.Flags=DXGI_SWAP_CHAIN_FLAG_ALLOW_MODE_SWITCH;
 IDXGISwapChain* swap=nullptr;ID3D11Device* device=nullptr;ID3D11DeviceContext* context=nullptr;
 if(FAILED(D3D11CreateDeviceAndSwapChain(nullptr,D3D_DRIVER_TYPE_HARDWARE,nullptr,0,nullptr,0,D3D11_SDK_VERSION,&desc,&swap,&device,nullptr,&context)))return 2;
 ShowWindow(window,SW_SHOW);SetForegroundWindow(window);
 if(exclusive&&FAILED(swap->SetFullscreenState(TRUE,nullptr)))return 3;
 ID3D11Texture2D* buffer=nullptr;ID3D11RenderTargetView* target=nullptr;swap->GetBuffer(0,__uuidof(ID3D11Texture2D),reinterpret_cast<void**>(&buffer));device->CreateRenderTargetView(buffer,nullptr,&target);buffer->Release();
 wchar_t mappingName[80];wsprintfW(mappingName,L"Local\\GemuOverlay-%lu",GetCurrentProcessId());HANDLE mapping=nullptr;Shared* shared=nullptr;
 DWORD started=GetTickCount();unsigned frames=0,overlayFrames=0,modeChanges=0,focusChanges=0,targetChanges=0;BOOL initialExclusive=FALSE;swap->GetFullscreenState(&initialExclusive,nullptr);
 DEVMODEW baseline{};baseline.dmSize=sizeof(baseline);EnumDisplaySettingsW(nullptr,ENUM_CURRENT_SETTINGS,&baseline);
 while(running&&GetTickCount()-started<12000){
  MSG message;while(PeekMessageW(&message,nullptr,0,0,PM_REMOVE)){TranslateMessage(&message);DispatchMessageW(&message);}
  if(!mapping){mapping=OpenFileMappingW(FILE_MAP_READ,FALSE,mappingName);if(mapping)shared=static_cast<Shared*>(MapViewOfFile(mapping,FILE_MAP_READ,0,0,sizeof(Shared)));}
  float seconds=(GetTickCount()-started)/1000.f;float color[]={.04f+.02f*std::sin(seconds),.1f,.16f+.03f*std::sin(seconds*2),1};
  context->OMSetRenderTargets(1,&target,nullptr);context->ClearRenderTargetView(target,color);swap->Present(1,0);frames++;
  ID3D11RenderTargetView* current=nullptr;context->OMGetRenderTargets(1,&current,nullptr);if(current!=target)targetChanges++;if(current)current->Release();
  if(shared&&read(shared->visible)){
   overlayFrames++;BOOL fullscreen=FALSE;swap->GetFullscreenState(&fullscreen,nullptr);if(fullscreen!=initialExclusive)modeChanges++;
   DEVMODEW active{};active.dmSize=sizeof(active);EnumDisplaySettingsW(nullptr,ENUM_CURRENT_SETTINGS,&active);
   if(active.dmPelsWidth!=baseline.dmPelsWidth||active.dmPelsHeight!=baseline.dmPelsHeight||active.dmDisplayFrequency!=baseline.dmDisplayFrequency)modeChanges++;
   if(GetForegroundWindow()!=window)focusChanges++;
  }
 }
 FILE* file=std::fopen(report,"w");if(file){std::fprintf(file,"{\"exclusive\":%s,\"frames\":%u,\"overlayFrames\":%u,\"modeChanges\":%u,\"focusChanges\":%u,\"targetChanges\":%u}\n",exclusive?"true":"false",frames,overlayFrames,modeChanges,focusChanges,targetChanges);std::fclose(file);}
 if(shared)UnmapViewOfFile(shared);if(mapping)CloseHandle(mapping);
 swap->SetFullscreenState(FALSE,nullptr);target->Release();context->Release();device->Release();swap->Release();DestroyWindow(window);
 return modeChanges||focusChanges||targetChanges||!overlayFrames?4:0;
}
