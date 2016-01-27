package windows

/*
#cgo CFLAGS: -DUNICODE -DWINVER=0x500

#include "win_wrapper.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

func OpenWindow(windowProc uintptr, width, height int) (windowHandle unsafe.Pointer, err error) {
	var window C.HWND
	errCode := C.openWindow(
		unsafe.Pointer(windowProc),
		C.int(width),
		C.int(height),
		&window,
	)
	if errCode == C.OK {
		return unsafe.Pointer(window), nil
	}
	if errCode == C.Error_RegisterClassEx {
		return nil, errors.New("OpenWindow: RegisterClassEx failed")
	}
	if errCode == C.Error_CreateWindowEx {
		return nil, errors.New("OpenWindow: CreateWindowEx failed")
	}
	return nil, errors.New("OpenWindow: unknown error")
}
