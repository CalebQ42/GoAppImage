package helpers

import (
	"debug/elf"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// This key in the desktop files written by appimaged describes where the AppImage is in the filesystem.
// We need this because we rewrite Exec= to include things like wrap and Firejail
const ExecLocationKey = "X-ExecLocation"

// CalculateElfSize returns the size of an ELF binary as an int64 based on the information in the ELF header
func CalculateElfSize(file string) int64 {

	// Open given elf file

	f, err := os.Open(file)
	PrintError("ioReader", err)
	// defer f.Close()
	if err != nil {
		return 0
	}

	_, err = f.Stat()
	PrintError("ioReader", err)
	if err != nil {
		return 0
	}

	e, err := elf.NewFile(f)
	if err != nil {
		PrintError("elfsize elf.NewFile", err)
		return 0
	}

	// Read identifier
	var ident [16]uint8
	_, err = f.ReadAt(ident[0:], 0)
	if err != nil {
		PrintError("elfsize read identifier", err)
		return 0
	}

	// Decode identifier
	if ident[0] != '\x7f' ||
		ident[1] != 'E' ||
		ident[2] != 'L' ||
		ident[3] != 'F' {
		log.Printf("Bad magic number at %d\n", ident[0:4])
		return 0
	}

	// Process by architecture
	sr := io.NewSectionReader(f, 0, 1<<63-1)
	var shoff, shentsize, shnum int64
	switch e.Class.String() {
	case "ELFCLASS64":
		hdr := new(elf.Header64)
		_, err = sr.Seek(0, 0)
		if err != nil {
			PrintError("elfsize", err)
			return 0
		}
		err = binary.Read(sr, e.ByteOrder, hdr)
		if err != nil {
			PrintError("elfsize", err)
			return 0
		}

		shoff = int64(hdr.Shoff)
		shnum = int64(hdr.Shnum)
		shentsize = int64(hdr.Shentsize)
	case "ELFCLASS32":
		hdr := new(elf.Header32)
		_, err = sr.Seek(0, 0)
		if err != nil {
			PrintError("elfsize", err)
			return 0
		}
		err = binary.Read(sr, e.ByteOrder, hdr)
		if err != nil {
			PrintError("elfsize", err)
			return 0
		}

		shoff = int64(hdr.Shoff)
		shnum = int64(hdr.Shnum)
		shentsize = int64(hdr.Shentsize)
	default:
		log.Println("unsupported elf architecture")
		return 0
	}

	// Calculate ELF size
	elfsize := shoff + (shentsize * shnum)
	// log.Println("elfsize:", elfsize, file)
	return elfsize
}

// Return true if magic string (hex) is found at offset
// TODO: Instead of magic string, could probably use something like []byte{'\r', '\n'} or []byte("AI")
func CheckMagicAtOffset(f *os.File, magic string, offset int64) bool {
	_, err := f.Seek(offset, 0) // Go to offset
	LogError("CheckMagicAtOffset: "+f.Name(), err)
	b := make([]byte, len(magic)/2) // Read bytes
	n, err := f.Read(b)
	LogError("CheckMagicAtOffset: "+f.Name(), err)
	hexmagic := hex.EncodeToString(b[:n])
	if hexmagic == magic {
		// if *verbosePtr == true {
		// 	log.Printf("CheckMagicAtOffset: %v: Magic 0x%x at offset %v\n", f.Name(), string(b[:n]), offset)
		// }
		return true
	}
	return false
}

// LogError logs error, prefixed by a string that explains the context
func LogError(context string, e error) {
	if e != nil {
		l := log.New(os.Stderr, "", 1)
		l.Println("ERROR " + context + ": " + e.Error())
	}
}

// PrintError prints error, prefixed by a string that explains the context
func PrintError(context string, e error) {
	if e != nil {
		os.Stderr.WriteString("ERROR " + context + ": " + e.Error() + "\n")
	}
}

// GetSectionData returns the contents of an ELF section and error
func GetSectionData(filepath string, name string) ([]byte, error) {
	// fmt.Println("GetSectionData for '" + name + "'")
	r, err := os.Open(filepath)
	if err == nil {
		defer r.Close()
	}
	f, err := elf.NewFile(r)
	if err != nil {
		return nil, err
	}
	section := f.Section(name)
	if section == nil {
		return nil, nil
	}
	data, err := section.Data()
	if err != nil {
		return nil, err
	}
	return data, nil
}

// FindMostRecentFile returns the most recent file
// from a slice of files, (currently) based on its mtime
// based on https://stackoverflow.com/a/45579190
// TODO: mtime may be fast, but is it "good enough" for our purposes?
func FindMostRecentFile(files []string) string {
	var modTime time.Time
	var names []string
	for _, f := range files {
		fi, _ := os.Stat(f)
		if fi.Mode().IsRegular() {
			if !fi.ModTime().Before(modTime) {
				if fi.ModTime().After(modTime) {
					modTime = fi.ModTime()
					names = names[:0]
				}
				names = append(names, f)
			}
		}
	}
	if len(names) > 0 {
		fmt.Println(modTime, names[0]) // Most recent
		return names[0]                // Most recent
	}
	return ""
}
