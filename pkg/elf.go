package pkg

import (
	"debug/elf"
	"errors"
	"fmt"
)

// Ldd - print shared object dependencies
func Ldd(path string) (x32 []string, x64 []string, err error) {

	// Open ELF file
	file, err := elf.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	x32, x64, err = DynString(file, elf.DT_NEEDED)
	if err != nil {
		return
	}
	return
}

// Copy pasted from debug/elf/elf.go

func stringTable(f *elf.File, link uint32) ([]byte, error) {
	if link <= 0 || link >= uint32(len(f.Sections)) {
		return nil, errors.New("section has invalid string table link")
	}
	return f.Sections[link].Data()
}

func getString(section []byte, start int) (string, bool) {
	if start < 0 || start >= len(section) {
		return "", false
	}

	for end := start; end < len(section); end++ {
		if section[end] == 0 {
			return string(section[start:end]), true
		}
	}
	return "", false
}

func DynString(f *elf.File, tag elf.DynTag) ([]string, []string, error) {
	switch tag {
	case elf.DT_NEEDED, elf.DT_SONAME, elf.DT_RPATH, elf.DT_RUNPATH:
	default:
		return nil, nil, fmt.Errorf("non-string-valued tag %v", tag)
	}
	ds := f.SectionByType(elf.SHT_DYNAMIC)
	if ds == nil {
		// not dynamic, so no libraries
		return nil, nil, nil
	}
	d, err := ds.Data()
	if err != nil {
		return nil, nil, err
	}
	str, err := stringTable(f, ds.Link)
	if err != nil {
		return nil, nil, err
	}
	var all32 []string
	var all64 []string

	for len(d) > 0 {
		var t elf.DynTag
		var v uint64
		switch f.Class {
		case elf.ELFCLASS32:
			t = elf.DynTag(f.ByteOrder.Uint32(d[0:4]))
			v = uint64(f.ByteOrder.Uint32(d[4:8]))
			d = d[8:]
		case elf.ELFCLASS64:
			t = elf.DynTag(f.ByteOrder.Uint64(d[0:8]))
			v = f.ByteOrder.Uint64(d[8:16])
			d = d[16:]
		}
		if t == tag {
			s, ok := getString(str, int(v))
			if ok {
				switch f.Class {
				case elf.ELFCLASS32:
					all32 = append(all32, s)
				case elf.ELFCLASS64:
					all64 = append(all64, s)
				}
			}
		}
	}
	return all32, all64, nil
}
