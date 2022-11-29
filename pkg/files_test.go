package pkg

import (
	"fmt"
	"testing"
)

// TestCheckElf - test checkElf function
func TestCheckElf(t *testing.T) {
	res, err := checkElf("/bin/ls")
	if err != nil {
		t.Error(err)
	}

	if res != true {
		t.Error("Expected true, got ", res)
	}
}

// TestFindElf - test FindElf function
func TestFindElf(t *testing.T) {
	res, err := FindElf("/home/platun0v")
	//res, err := FindElf("/home/platun0v/.cache/yay/bitcoin-core/pkg")

	if err != nil {
		t.Error(err)
	}

	fmt.Println(res)
	fmt.Println(len(res))

	if len(res) == 0 {
		t.Error("Expected not empty array, got ", res)
	}
}
