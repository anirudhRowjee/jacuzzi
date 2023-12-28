package jacuzzi

import (
	"os"
	"testing"
)

func TestEverythingWorks(t *testing.T) {
	if (1 + 1) != 2 {
		t.Error("Addition is not working!")
	}
}

// test to see if the pagecache will initialize
func TestInitializePageCache(t *testing.T) {

	pageSize := os.Getpagesize()
	slots := 10

	var pc PageCache = PageCache{}
	pc.Init(pageSize, slots)

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
