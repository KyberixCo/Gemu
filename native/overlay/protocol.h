#pragma once
#include <windows.h>
#include <cstdint>

struct Shared {
 uint32_t version, size;
 volatile LONG ready, visible, command, state, language;
 volatile LONG hostTick, rendererTick, stop, width, height;
 char title[512];
};
static_assert(sizeof(Shared)==560,"IPC layout must match Go Shared");
inline LONG read(volatile LONG& value) { return InterlockedCompareExchange(&value,0,0); }
inline void write(volatile LONG& value,LONG next) { InterlockedExchange(&value,next); }
