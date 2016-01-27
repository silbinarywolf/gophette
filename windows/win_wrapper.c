#include "win_wrapper.h"

int openWindow(void* proc, int width, int height, HWND* window) {
	WNDCLASSEX cls = {0};
	cls.cbSize = sizeof(WNDCLASSEX);
	cls.lpfnWndProc = (WNDPROC)proc;
	cls.lpszClassName = L"GoWindowClass";
	cls.hCursor = LoadCursor(0, IDC_ARROW);

	ATOM atom = RegisterClassEx(&cls);
	if (atom == 0) {
		return Error_RegisterClassEx;
	}

	*window = CreateWindowEx(0, L"GoWindowClass", "",
		WS_OVERLAPPEDWINDOW | WS_VISIBLE,
		100, 100, width, height, 0, 0, 0, 0);
	if (*window == 0) {
		return Error_CreateWindowEx;
	}

	return OK;
}
