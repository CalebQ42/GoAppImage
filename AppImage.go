package goappimage

/*
#cgo CFLAGS: -I/usr/lib
#cgo LDFLAGS: -L/usr/lib/libappimage.so -lappimage

#include <appimage/appimage.h>
#include <stdlib.h>

int char_length(char** in){
	int i = 0;
	for (; in[i] != NULL ; i++);
	return i;
}
*/
import "C"

import (
	"fmt"
	"strings"
	"unsafe"
)

//AppImage is the representation of an AppImage. You must call Free() once your done to manually free up the C variables.
//It's recommended to defer Free() immediately after declaring so you don't forget.
type AppImage struct {
	InternalFiles        []string
	location             string
	clocation            *C.char
	cinternalFiles       **C.char
	desktopFileLocation  string
	cdesktopFileLocation *C.char
}

//Free manually frees memory alocated to AppImage's C variables.
func (a *AppImage) Free() {
	C.free(unsafe.Pointer(a.clocation))
	C.free(unsafe.Pointer(a.cdesktopFileLocation))
	C.appimage_string_list_free(a.cinternalFiles)
}

//NewAppImage creates a new AppImage tied to location.
func NewAppImage(location string) AppImage {
	if strings.HasSuffix(location, ".AppImage") {
		fmt.Println("The given location does not appear to be a an AppImage, this may cause issues with many things")
	}
	var out AppImage
	out.location = location
	out.clocation = C.CString(out.location)
	out.cinternalFiles = C.appimage_list_files(out.clocation)
	cfilesLength := C.char_length(out.cinternalFiles)
	tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(out.cinternalFiles))[:cfilesLength:cfilesLength]
	out.InternalFiles = make([]string, cfilesLength)
	for i, v := range tmpslice {
		tmp := C.GoString(v)
		out.InternalFiles[i] = tmp
		if strings.HasSuffix(tmp, ".desktop") {
			out.desktopFileLocation = tmp
			out.cdesktopFileLocation = v
		}
	}
	return out
}

//ExtractFile extracts the file at location to extractLocation. File should be found in AppImage.InternalFiles.
func (a *AppImage) ExtractFile(location, extractLocation string) {
	cloc := C.CString(location)
	cextract := C.CString(extractLocation)
	defer C.free(unsafe.Pointer(cloc))
	defer C.free(unsafe.Pointer(cextract))
	C.appimage_extract_file_following_symlinks(a.clocation, cloc, cextract)
}

//ExtractDesktop extracts the desktop file to extractLocation.
func (a *AppImage) ExtractDesktop(extractLocation string) {
	cextract := C.CString(extractLocation)
	defer C.free(unsafe.Pointer(cextract))
	C.appimage_extract_file_following_symlinks(a.clocation, a.cdesktopFileLocation, cextract)
}
