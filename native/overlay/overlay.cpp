#include "protocol.h"
#include <d3d11.h>
#include <dxgi.h>
#include <xinput.h>
#include <algorithm>
#include <cmath>
#include "MinHook.h"
#include "imgui.h"
#include "imgui_impl_dx11.h"
#include "imgui_impl_win32.h"

using PresentFn=HRESULT(STDMETHODCALLTYPE*)(IDXGISwapChain*,UINT,UINT);
using ResizeFn=HRESULT(STDMETHODCALLTYPE*)(IDXGISwapChain*,UINT,UINT,UINT,DXGI_FORMAT,UINT);
using InputFn=DWORD(WINAPI*)(DWORD,XINPUT_STATE*);
using KeyFn=SHORT(WINAPI*)(int);
static PresentFn originalPresent;
static ResizeFn originalResize;
static InputFn originalInput[3]{};
static KeyFn originalKey;
static Shared* shared;
static HANDLE mapping;
static SRWLOCK renderLock=SRWLOCK_INIT;
static IDXGISwapChain* chain;
static ID3D11Device* device;
static ID3D11DeviceContext* context;
static ID3D11RenderTargetView* target;
static HWND gameWindow;
static WNDPROC originalWndProc;
static bool confirmation=false, wasVisible=false;
static int selected=0;
static DWORD lastPadAction=0;
static volatile LONG releaseGate=0;

static bool visible(){return shared&&read(shared->stop)==0&&DWORD(GetTickCount()-DWORD(read(shared->hostTick)))<3000&&read(shared->visible)!=0;}
static bool blocked(){return visible()||read(releaseGate)!=0;}
extern IMGUI_IMPL_API LRESULT ImGui_ImplWin32_WndProcHandler(HWND,UINT,WPARAM,LPARAM);

static LRESULT CALLBACK windowProc(HWND hwnd,UINT message,WPARAM wp,LPARAM lp){
 if(blocked()){
  AcquireSRWLockExclusive(&renderLock);
  if(ImGui::GetCurrentContext())ImGui_ImplWin32_WndProcHandler(hwnd,message,wp,lp);
  ReleaseSRWLockExclusive(&renderLock);
  // Consume standard keyboard/mouse and raw input only while the menu is open.
  if((message>=WM_KEYFIRST&&message<=WM_KEYLAST)||(message>=WM_MOUSEFIRST&&message<=WM_MOUSELAST)||message==WM_INPUT)
   return DefWindowProcW(hwnd,message,wp,lp);
 }
 return CallWindowProcW(originalWndProc,hwnd,message,wp,lp);
}
static SHORT WINAPI keyState(int key){return blocked()?0:originalKey(key);}
static DWORD inputState(int api,DWORD index,XINPUT_STATE* state){
 DWORD result=originalInput[api](index,state);
 if(result==ERROR_SUCCESS&&blocked())state->Gamepad={};
 return result;
}
static DWORD WINAPI input0(DWORD slot,XINPUT_STATE* state){return inputState(0,slot,state);}
static DWORD WINAPI input1(DWORD slot,XINPUT_STATE* state){return inputState(1,slot,state);}
static DWORD WINAPI input2(DWORD slot,XINPUT_STATE* state){return inputState(2,slot,state);}

static void send(LONG command){
 if(InterlockedCompareExchange(&shared->command,command,0)==0){write(releaseGate,1);write(shared->visible,0);confirmation=false;selected=0;}
}
static const char* text(const char* es,const char* en){return read(shared->language)==1?en:es;}
static void navigation(bool& accept,bool& back,int count){
 auto& io=ImGui::GetIO();
 bool up=ImGui::IsKeyPressed(ImGuiKey_UpArrow,true),down=ImGui::IsKeyPressed(ImGuiKey_DownArrow,true);
 accept=ImGui::IsKeyPressed(ImGuiKey_Enter,false)||ImGui::IsKeyPressed(ImGuiKey_Space,false);
 back=ImGui::IsKeyPressed(ImGuiKey_Escape,false);
 // Use the original XInput entry point, bypassing the game's neutral input.
 bool padHeld=false;
 for(auto api:originalInput){
  if(!api)continue;
  for(DWORD slot=0;slot<4;slot++){
   XINPUT_STATE state{};if(api(slot,&state)!=ERROR_SUCCESS)continue;
   auto buttons=state.Gamepad.wButtons;
   int direction=(buttons&XINPUT_GAMEPAD_DPAD_UP||state.Gamepad.sThumbLY>16000)?-1:(buttons&XINPUT_GAMEPAD_DPAD_DOWN||state.Gamepad.sThumbLY<-16000)?1:0;
   DWORD action=direction==-1?1:direction==1?2:buttons&XINPUT_GAMEPAD_A?3:buttons&XINPUT_GAMEPAD_B?4:0;
   if(action){padHeld=true;if(GetTickCount()-lastPadAction>220){up|=action==1;down|=action==2;accept|=action==3;back|=action==4;lastPadAction=GetTickCount();}}
   break;
  }
  break;
 }
 if(!padHeld)lastPadAction=0;
 if(up)selected=(selected+count-1)%count;
 if(down)selected=(selected+1)%count;
 // Tab follows the same vertical action order.
 if(ImGui::IsKeyPressed(ImGuiKey_Tab,true))selected=(selected+(io.KeyShift?count-1:1))%count;
}

static void drawMenu(){
 auto& io=ImGui::GetIO();
 float scale=std::clamp(io.DisplaySize.y/720.f,.6f,3.f);
 io.FontGlobalScale=1.15f*scale;
 const float gutter=50*scale,width=std::min(500*scale,io.DisplaySize.x-gutter*2);
 ImGui::SetNextWindowPos({0,0});ImGui::SetNextWindowSize(io.DisplaySize);
 ImGui::Begin("Gemu",nullptr,ImGuiWindowFlags_NoDecoration|ImGuiWindowFlags_NoMove|ImGuiWindowFlags_NoSavedSettings|ImGuiWindowFlags_NoBackground|ImGuiWindowFlags_NoBringToFrontOnFocus);
 auto draw=ImGui::GetWindowDrawList();auto font=ImGui::GetFont();
 const ImU32 acid=IM_COL32(217,255,63,255),paper=IM_COL32(242,239,230,255),muted=IM_COL32(177,185,186,255);
 draw->AddRectFilled({0,0},io.DisplaySize,IM_COL32(8,12,18,220));
 draw->AddText(font,32*scale,{gutter,28*scale},paper,"GEMU");
 draw->AddText(font,13*scale,{gutter+140*scale,42*scale},acid,text("PARTIDA ACTUAL","CURRENT GAME"));
 draw->AddLine({gutter,85*scale},{io.DisplaySize.x-gutter,85*scale},IM_COL32(255,255,255,25));
 draw->AddText(font,30*scale,{gutter,125*scale},paper,confirmation?text("¿Cerrar el juego?","Close the game?"):text("Tu partida","Your game"));
 if(confirmation)draw->AddText(font,15*scale,{gutter,176*scale},IM_COL32(255,208,182,255),text("El progreso sin guardar puede perderse.","Unsaved progress may be lost."));
 int count=confirmation?2:3;bool accept=false,back=false;navigation(accept,back,count);
 if(back){if(confirmation){confirmation=false;selected=2;}else send(1);}
 float startY=(confirmation?220:195)*scale;
 const char* labels[]={text("Continuar jugando","Resume game"),text("Volver a Gemu","Return to Gemu"),text("Cerrar juego","Close game")};
 const char* descriptions[]={text("Regresar directamente al juego.","Go straight back to the game."),text("Explorar la biblioteca sin cerrar el juego.","Browse the library while the game keeps running."),text("Terminar la partida.","End this game session.")};
 if(confirmation){labels[0]=text("Cancelar","Cancel");labels[1]=text("Confirmar cierre","Confirm close");descriptions[0]=text("Volver a las opciones de la partida.","Return to the game menu.");descriptions[1]=text("Solicitar el cierre del juego.","Request that the game closes.");}
 for(int i=0;i<count;i++){
  float y=startY+i*98*scale+(i==2?14*scale:0);
  ImGui::SetCursorPos({gutter,y});ImGui::PushID(i);
  bool enabled=confirmation?i==0||read(shared->state)==2:i!=2||read(shared->state)==2;
  bool clicked=ImGui::InvisibleButton("action",{width,86*scale});
  if(ImGui::IsItemHovered()&&(std::abs(io.MouseDelta.x)+std::abs(io.MouseDelta.y)>0))selected=i;
  bool focused=selected==i;
  ImU32 bg=i==0?acid:focused?IM_COL32(217,255,63,16):IM_COL32(255,255,255,5);
  draw->AddRectFilled({gutter,y},{gutter+width,y+86*scale},bg,0);
  if(focused)draw->AddRectFilled({gutter,y},{gutter+3*scale,y+86*scale},i==0?IM_COL32(20,27,8,255):acid);
  ImU32 color=i==0?IM_COL32(20,27,8,255):enabled?paper:muted;
  draw->AddText(font,21*scale,{gutter+24*scale,y+17*scale},color,labels[i]);
  draw->AddText(font,13*scale,{gutter+24*scale,y+49*scale},i==0?IM_COL32(52,64,20,255):muted,descriptions[i]);
  if(enabled&&(clicked||(accept&&focused))){
   if(confirmation){if(i==0){confirmation=false;selected=2;}else send(3);}
   else if(i==0)send(1);else if(i==1)send(2);else{confirmation=true;selected=0;}
   ImGui::PopID();break;
  }
  ImGui::PopID();
 }
 if(io.DisplaySize.x>1050*scale){
  float x=gutter+width+70*scale;
  draw->AddText(font,16*scale,{x,230*scale},acid,text("EN EJECUCIÓN","RUNNING"));
  draw->AddText(font,34*scale,{x,275*scale},paper,shared->title,nullptr,io.DisplaySize.x-x-gutter);
  draw->AddText(font,14*scale,{x,440*scale},muted,text("El juego sigue ejecutándose.","The game keeps running."));
 }
 float footer=io.DisplaySize.y-70*scale;
 draw->AddLine({gutter,footer},{io.DisplaySize.x-gutter,footer},IM_COL32(255,255,255,25));
 draw->AddText(font,13*scale,{gutter,footer+25*scale},muted,text("↑ ↓ / Cruceta   Navegar       Enter / A   Seleccionar       Esc / B   Volver       F10   Menú","↑ ↓ / D-pad   Navigate       Enter / A   Select       Esc / B   Back       F10   Menu"));
 ImGui::End();
}

static bool initialise(IDXGISwapChain* swap){
 DXGI_SWAP_CHAIN_DESC desc{};if(FAILED(swap->GetDesc(&desc))||desc.BufferDesc.Width<640||desc.BufferDesc.Height<360)return false;
 if(FAILED(swap->GetDevice(__uuidof(ID3D11Device),reinterpret_cast<void**>(&device))))return false;
 device->GetImmediateContext(&context);gameWindow=desc.OutputWindow;chain=swap;
 IMGUI_CHECKVERSION();ImGui::CreateContext();auto& io=ImGui::GetIO();io.IniFilename=nullptr;io.LogFilename=nullptr;
 auto& style=ImGui::GetStyle();style.WindowRounding=style.FrameRounding=style.PopupRounding=style.TabRounding=0;style.WindowPadding={0,0};
 if(!ImGui_ImplWin32_Init(gameWindow)||!ImGui_ImplDX11_Init(device,context))return false;
 originalWndProc=reinterpret_cast<WNDPROC>(SetWindowLongPtrW(gameWindow,GWLP_WNDPROC,reinterpret_cast<LONG_PTR>(windowProc)));
 return originalWndProc!=nullptr;
}
static HRESULT STDMETHODCALLTYPE present(IDXGISwapChain* swap,UINT interval,UINT flags){
 if(shared&&read(shared->stop)==0&&!(flags&DXGI_PRESENT_TEST)){
  AcquireSRWLockExclusive(&renderLock);
  if(!chain)initialise(swap);
  if(chain==swap&&context&&ImGui::GetCurrentContext()){
   DXGI_SWAP_CHAIN_DESC desc{};swap->GetDesc(&desc);
   write(shared->width,desc.BufferDesc.Width);write(shared->height,desc.BufferDesc.Height);
   write(shared->rendererTick,GetTickCount());write(shared->ready,1);
   bool open=visible();if(open&&!wasVisible){confirmation=false;selected=0;lastPadAction=GetTickCount();}
   if(!open&&wasVisible)write(releaseGate,1);
   if(!open&&read(releaseGate)){
    bool held=originalKey&&((originalKey(VK_RETURN)|originalKey(VK_SPACE)|originalKey(VK_ESCAPE))&0x8000);
    for(auto api:originalInput){if(!api)continue;for(DWORD slot=0;slot<4;slot++){XINPUT_STATE state{};if(api(slot,&state)==ERROR_SUCCESS)held|=(state.Gamepad.wButtons&(XINPUT_GAMEPAD_A|XINPUT_GAMEPAD_B))!=0;}}
    if(!held)write(releaseGate,0);
   }
   wasVisible=open;ImGui::GetIO().MouseDrawCursor=open;
   if(open){
    if(!target){ID3D11Texture2D* buffer=nullptr;if(SUCCEEDED(swap->GetBuffer(0,__uuidof(ID3D11Texture2D),reinterpret_cast<void**>(&buffer)))){device->CreateRenderTargetView(buffer,nullptr,&target);buffer->Release();}}
    if(target){
     ImGui_ImplDX11_NewFrame();ImGui_ImplWin32_NewFrame();ImGui::NewFrame();drawMenu();ImGui::Render();
     ID3D11RenderTargetView* saved[D3D11_SIMULTANEOUS_RENDER_TARGET_COUNT]{};ID3D11DepthStencilView* depth=nullptr;
     context->OMGetRenderTargets(D3D11_SIMULTANEOUS_RENDER_TARGET_COUNT,saved,&depth);
     context->OMSetRenderTargets(1,&target,nullptr);ImGui_ImplDX11_RenderDrawData(ImGui::GetDrawData());
     context->OMSetRenderTargets(D3D11_SIMULTANEOUS_RENDER_TARGET_COUNT,saved,depth);
     for(auto view:saved)if(view)view->Release();if(depth)depth->Release();
    }
   }
  }
  ReleaseSRWLockExclusive(&renderLock);
 }
 return originalPresent(swap,interval,flags);
}
static HRESULT STDMETHODCALLTYPE resize(IDXGISwapChain* swap,UINT count,UINT width,UINT height,DXGI_FORMAT format,UINT flags){
 AcquireSRWLockExclusive(&renderLock);
 if(swap==chain&&target){target->Release();target=nullptr;}
 auto result=originalResize(swap,count,width,height,format,flags);
 ReleaseSRWLockExclusive(&renderLock);return result;
}

static DWORD WINAPI bootstrap(void*){
 wchar_t name[80];wsprintfW(name,L"Local\\GemuOverlay-%lu",GetCurrentProcessId());
 mapping=OpenFileMappingW(FILE_MAP_READ|FILE_MAP_WRITE,FALSE,name);if(!mapping)return 1;
 shared=static_cast<Shared*>(MapViewOfFile(mapping,FILE_MAP_READ|FILE_MAP_WRITE,0,0,sizeof(Shared)));
 if(!shared||shared->version!=1||shared->size!=sizeof(Shared)||read(shared->stop))return 2;
 HWND dummy=CreateWindowExW(0,L"STATIC",L"Gemu hook bootstrap",WS_POPUP,0,0,2,2,nullptr,nullptr,GetModuleHandleW(nullptr),nullptr);
 if(!dummy)return 3;
 DXGI_SWAP_CHAIN_DESC desc{};desc.BufferCount=1;desc.BufferDesc.Width=2;desc.BufferDesc.Height=2;desc.BufferDesc.Format=DXGI_FORMAT_R8G8B8A8_UNORM;desc.BufferUsage=DXGI_USAGE_RENDER_TARGET_OUTPUT;desc.OutputWindow=dummy;desc.SampleDesc.Count=1;desc.Windowed=TRUE;
 IDXGISwapChain* swap=nullptr;ID3D11Device* sampleDevice=nullptr;ID3D11DeviceContext* sampleContext=nullptr;
 HRESULT result=D3D11CreateDeviceAndSwapChain(nullptr,D3D_DRIVER_TYPE_HARDWARE,nullptr,0,nullptr,0,D3D11_SDK_VERSION,&desc,&swap,&sampleDevice,nullptr,&sampleContext);
 if(FAILED(result)){DestroyWindow(dummy);return 4;}
 auto vtable=*reinterpret_cast<void***>(swap);void* p=vtable[8];void* r=vtable[13];
 sampleContext->Release();sampleDevice->Release();swap->Release();DestroyWindow(dummy);
 if(MH_Initialize()!=MH_OK)return 5;
 if(MH_CreateHook(p,reinterpret_cast<void*>(present),reinterpret_cast<void**>(&originalPresent))!=MH_OK||MH_CreateHook(r,reinterpret_cast<void*>(resize),reinterpret_cast<void**>(&originalResize))!=MH_OK)return 6;
 auto key=GetProcAddress(GetModuleHandleW(L"user32.dll"),"GetAsyncKeyState");
 if(key)MH_CreateHook(reinterpret_cast<void*>(key),reinterpret_cast<void*>(keyState),reinterpret_cast<void**>(&originalKey));
 const wchar_t* libraries[]={L"xinput1_4.dll",L"xinput1_3.dll",L"xinput9_1_0.dll"};void* hooks[]={reinterpret_cast<void*>(input0),reinterpret_cast<void*>(input1),reinterpret_cast<void*>(input2)};
 for(int i=0;i<3;i++){auto library=GetModuleHandleW(libraries[i]);if(!library)library=LoadLibraryExW(libraries[i],nullptr,LOAD_LIBRARY_SEARCH_SYSTEM32);if(!library)continue;auto proc=GetProcAddress(library,"XInputGetState");if(proc)MH_CreateHook(reinterpret_cast<void*>(proc),hooks[i],reinterpret_cast<void**>(&originalInput[i]));}
 return MH_EnableHook(MH_ALL_HOOKS)==MH_OK?0:7;
}
BOOL WINAPI DllMain(HINSTANCE module,DWORD reason,void*){
 if(reason==DLL_PROCESS_ATTACH){DisableThreadLibraryCalls(module);auto thread=CreateThread(nullptr,0,bootstrap,nullptr,0,nullptr);if(thread)CloseHandle(thread);}
 return TRUE;
}
