#include <Windows.h>

#define OK                     0
#define Error_RegisterClassEx -1
#define Error_CreateWindowEx  -2

int openWindow(void* proc, int width, int height, HWND* window);