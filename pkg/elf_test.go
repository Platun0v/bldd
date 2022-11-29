package pkg

import (
	"reflect"
	"testing"
)

func TestLdd(t *testing.T) {
	_, x64, err := Ldd("/bin/ls")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(x64, []string{"libcap.so.2", "libc.so.6"}) {
		t.Error("x64 libraries are not correct")
	}
}
