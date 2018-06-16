package goappimage

/*
#cgo CFLAGS: -I${SRCDIR}/AppImageKit/
#cgo LDFLAGS: -L${SRCDIR}/AppImageKit -lappimage

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
	//InternalFiles is a slice containing the names of ALL the AppImage's files. Backed by a C array.
	InternalFiles        []string
	cinternalFiles       **C.char
	location             string
	clocation            *C.char
	desktopFileLocation  string
	cdesktopFileLocation *C.char
	cmd5                 *C.char
	initialized          bool
	md5ed                bool
}

//Free manually frees memory alocated to AppImage's C variables.
func (a *AppImage) Free() {
	C.free(unsafe.Pointer(a.clocation))
	if a.initialized {
		C.appimage_string_list_free(a.cinternalFiles)
	}
	if a.md5ed {
		C.free(unsafe.Pointer(a.cmd5))
	}
}

//NewAppImage creates a new AppImage tied to location.
func NewAppImage(location string) AppImage {
	if !strings.HasSuffix(location, ".AppImage") {
		fmt.Println("The given location does not appear to be a an AppImage, this may cause issues with many things")
	}
	var out AppImage
	out.location = location
	out.clocation = C.CString(out.location)
	return out
}

//Initialize is a long process that allows some AppImage functions to work. Takes a long time.
func (a *AppImage) Initialize() {
	a.cinternalFiles = C.appimage_list_files(a.clocation)
	cfilesLength := C.char_length(a.cinternalFiles)
	tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(a.cinternalFiles))[:cfilesLength:cfilesLength]
	a.InternalFiles = make([]string, cfilesLength)
	for i, v := range tmpslice {
		tmp := C.GoString(v)
		a.InternalFiles[i] = tmp
		if strings.HasSuffix(tmp, ".desktop") {
			a.desktopFileLocation = tmp
			a.cdesktopFileLocation = v
		}
	}
	a.initialized = true
}

//Md5 returns the md5 hash of the appimage
func (a *AppImage) Md5() string {
	a.cmd5 = C.appimage_get_md5(a.clocation)
	a.md5ed = true
	return C.GoString(a.cmd5)
}

//ExtractFile extracts the file at location to extractLocation. File should be found in AppImage.InternalFiles.
func (a *AppImage) ExtractFile(location, extractLocation string) {
	cloc := C.CString(location)
	cextract := C.CString(extractLocation)
	defer C.free(unsafe.Pointer(cloc))
	defer C.free(unsafe.Pointer(cextract))
	C.appimage_extract_file_following_symlinks(a.clocation, cloc, cextract)
}

//ExtractDesktop extracts the desktop file to extractLocation. Requires initialization.
func (a *AppImage) ExtractDesktop(extractLocation string) {
	if a.initialized {
		cextract := C.CString(extractLocation)
		defer C.free(unsafe.Pointer(cextract))
		C.appimage_extract_file_following_symlinks(a.clocation, a.cdesktopFileLocation, cextract)
	} else {
		fmt.Println("AppImage needs to be initialized before this works")
	}
}
