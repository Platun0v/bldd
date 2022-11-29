package pkg

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

// FindElf - find all ELF files in specified directory. If passed file, check if it is ELF.
func FindElf(path string) ([]string, error) {
	var res []string
	items, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) { // skip permission denied
			fmt.Println("Permission denied: " + path)
			return res, nil
		} else if os.IsNotExist(err) { // skip not found
			fmt.Println("Not found: " + path)
			return res, nil
		} else if errors.Is(err, syscall.ENOTDIR) { // if not directory, check if it is ELF
			isElf, err := checkElf(path)
			if err != nil {
				return res, fmt.Errorf("error while checking file %s: %s", path, err)
			}
			if isElf {
				return []string{path}, nil
			}
		}

		return res, err
	}

	for _, item := range items {
		if item.IsDir() {
			dirRes, err := FindElf(path + "/" + item.Name())
			if err != nil {
				return res, err
			}
			res = append(res, dirRes...)
		} else {
			if item.Type().IsRegular() {
				isElf, err := checkElf(path + "/" + item.Name())
				if err != nil {
					return res, fmt.Errorf("error while checking file %s: %s", path+"/"+item.Name(), err)
				}
				if isElf {
					res = append(res, path+"/"+item.Name())
				}
			}
		}
	}
	return res, nil
}

// checkElf - check if file is ELF
func checkElf(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) { // skip permission denied
			fmt.Println("Permission denied: " + path)
			return false, nil
		}
		return false, err
	}
	defer file.Close()

	// ELF header
	elfHeader := []byte{0x7f, 'E', 'L', 'F'}
	header := make([]byte, 4)
	_, err = file.Read(header)
	switch err {
	case nil:
	case io.EOF: // empty file
		return false, nil
	default:
		return false, err
	}

	if bytes.Equal(header, elfHeader) {
		return true, nil
	}
	return false, nil
}
