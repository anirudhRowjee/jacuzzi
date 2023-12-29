package jacuzzi

import (
	"fmt"
	"os"
)

/*
# Jacuzzi - A concurrent Buffer Pool / Page Cache for Go

A simple page cache for golang
- Replacement Policy TBD
- Configurable Page Size
- Configurable Slot Count
- Encourages Concurrent Use with fine-grained locking
*/

type Frame struct {
	frameId int
	PageId  int
	Pinned  bool
	Dirty   bool
	data    []byte
}

// Initialize the frame by allocating the data byte array
func (f *Frame) Initialize(p *PageCache, frameId int) {
	f.frameId = frameId
	f.data = make([]byte, p.PageSize)
	f.Dirty = false
	f.Pinned = false
	f.PageId = 0
}

// Reset a frame to reuse it
func (f *Frame) Reset() {
	f.frameId = 0
	f.Dirty = false
	f.Pinned = false
	f.PageId = 0
}

type PageCache struct {
	PageSize  int         // const: number of bytes in a page, loaded at runtime
	Data      map[int]int // Data is the mapping of address to frames
	Frames    []Frame     // Frames is where all page frames live, never get deleted
	SlotCount int         // the number of slots this pagecache has
	file      *os.File    // backing file
}

// Start the pagecache
func (p *PageCache) Init(PageSize, slots int, filename string) error {
	p.PageSize = PageSize
	p.SlotCount = slots

	// Open the File
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 066)
	if err != nil {
		return err
	}
	p.file = file

	// Allocate all Maps
	p.Data = make(map[int]int, p.SlotCount) // TODO see if we need this to be constant-sized
	p.Frames = make([]Frame, p.SlotCount)

	// Allocate all Frames
	for idx := range p.Frames {
		p.Frames[idx].Initialize(p, idx)
	}
	return nil
}

func CheckMapKeyPresent(mymap *map[int]int, target int) bool {
	for k, _ := range *mymap {
		if k == target {
			return true
		}
	}
	return false
}

// A function that reads a page into a buffer
// return signature (bool, error) ->
// ----> bool: True if Cache Hit, False if Cache Miss
func (p *PageCache) ReadPage(PageStartAddress int, Destination []byte) (bool, error) {

	// Check if the offset exists in the translation table
	hit := CheckMapKeyPresent(&p.Data, PageStartAddress)

	if hit {
		// If it exists, copy the page into the destination
		TargetFrameId := p.Data[PageStartAddress]
		BytesCopied := copy(Destination, p.Frames[TargetFrameId].data)
		if BytesCopied != p.PageSize {
			return hit, fmt.Errorf(
				"Could not Copy a Full Page -> Expected %d Copied %d",
				p.PageSize,
				BytesCopied,
			)
		}

	} else {
		// If it doesn't exist, evict a page, and fill the frame with the part necessary
		// As of right now, this just returns the first page
		NewFrameId := p.EvictPage()
		// Reset the new Frame
		p.Frames[NewFrameId].Reset()
		// Copy the data from the file into the new frame
		bytes_read, err := p.file.ReadAt(p.Frames[NewFrameId].data, int64(PageStartAddress))
		if err != nil {
			return hit, err
		}
		if bytes_read != p.PageSize {
			return hit, fmt.Errorf(
				"Could not Copy a Full Page -> Expected %d Copied %d",
				p.PageSize,
				bytes_read,
			)
		}
		// Add the mapping of the offset into the data table
		p.Data[PageStartAddress] = NewFrameId
		// Finally, read into target buffer
		BytesCopied := copy(Destination, p.Frames[NewFrameId].data)
		if BytesCopied != p.PageSize {
			return hit, fmt.Errorf(
				"Could not Copy a Full Page -> Expected %d Copied %d",
				p.PageSize,
				BytesCopied,
			)
		}
	}
	return hit, nil
}

// A function that writes a page from the application to the pagecache
// return signature (bool, error) ->
// ----> bool: True if Cache Hit, False if Cache Miss
func (p *PageCache) WritePage(PageStartAddress int, Content *[]byte) (bool, error) {
	return false, nil
}

// A function that writes a page frame from the cache to the disk
func (p *PageCache) FlushFrameToDisk(FrameID int) (bool, error) {

	return false, nil
}

// function to evict page and return pageId of new empty frame
func (p *PageCache) EvictPage() int {

	return 0
}

// TODO
// Add functions for
// -> Flushing a Page to Disk
// -> Evicting a Page from the Cache
// ->
