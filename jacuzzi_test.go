package jacuzzi

import (
	"bytes"
	"os"
	"testing"
)

// test to see if the pagecache will initialize
func TestInitializePageCache(t *testing.T) {

	pageSize := os.Getpagesize()
	slots := 10
	file := "test.bin"

	var pc PageCache = PageCache{}
	err := pc.Init(pageSize, slots, file)
	if err != nil {
		t.Error(err)
	}

	// Check if all the frames have been initialized
	if pc.SlotCount != slots {
		t.Errorf(
			"Could not Initialize Slot Counter properly: Expected %d got %d\n",
			slots,
			pc.SlotCount,
		)
	}

	// check if there are `slots` entries present in the frames array
	if len(pc.Frames) != pc.SlotCount {
		t.Errorf(
			"Could not Initialize Frames properly: Expected %d got %d\n",
			len(pc.Frames),
			pc.SlotCount,
		)
	}
}

// Check if reading into the pagecache works
func TestReadIntoCache(t *testing.T) {

	pageSize := os.Getpagesize()
	slots := 10
	file := "test.bin"

	var pc PageCache = PageCache{}
	err := pc.Init(pageSize, slots, file)
	if err != nil {
		t.Error(err)
	}

	// Read the first page of the file into memory
	buffer1 := make([]byte, pageSize)
	hit1, err := pc.ReadPage(0, buffer1)
	if err != nil {
		t.Error(err)
	}

	buffer2 := make([]byte, pageSize)
	hit2, err := pc.ReadPage(0, buffer2)
	if err != nil {
		t.Error(err)
	}

	if hit1 != false {
		t.Error("Unexpected Cache Hit on first read of page, hit status: ", hit1)
	}
	if hit2 != true {
		t.Error("Unexpected Cache Miss on second read of page, hit status: ", hit2)
	}
	if bytes.Equal(buffer1, buffer2) != true {
		t.Errorf("Data Drift in bytes read from cache: Inconsistency between first and second read")
	}
}
