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
	frameId          int // ID local to the frame; Never leaves the frame boundary
	PageOffsetInFile int // Offset in the file where this page exists
	Pinned           bool
	Dirty            bool
	ReferenceCount   int
	data             []byte
}

// Initialize the frame by allocating the data byte array
func (f *Frame) Initialize(p *PageCache, frameId int) {
	f.frameId = frameId
	f.data = make([]byte, p.PageSize)
	f.Dirty = false
	f.Pinned = false
	f.PageOffsetInFile = 0
	f.ReferenceCount = 0
}

// Reset a frame to reuse it
func (f *Frame) Reset() {
	f.frameId = 0
	f.Dirty = false
	f.Pinned = false
	f.ReferenceCount = 0
	f.PageOffsetInFile = 0
}

type PageCache struct {
	PageSize   int         // const: number of bytes in a page, loaded at runtime
	Data       map[int]int // Data is the mapping of address to frames
	Frames     []Frame     // Frames is where all page frames live, never get deleted
	FrameCount int         // the number of slots this pagecache has
	file       *os.File    // backing file
}

// Start the pagecache
func (p *PageCache) Init(PageSize, slots int, filename string) error {
	p.PageSize = PageSize
	p.FrameCount = slots

	// Open the File
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 066)
	if err != nil {
		return err
	}
	p.file = file

	// Allocate all Maps
	p.Data = make(map[int]int, p.FrameCount) // TODO see if we need this to be constant-sized
	p.Frames = make([]Frame, p.FrameCount)

	// Allocate all Frames
	for idx := range p.Frames {
		p.Frames[idx].Initialize(p, idx)
	}
	return nil
}

func CheckMapKeyPresent(mymap *map[int]int, target int) bool {
	for k := range *mymap {
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
		// NOTE all pages that are read are referenced by default; one must manually dereference them to enable eviction
		p.Frames[TargetFrameId].ReferenceCount += 1
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
		NewFrameId, err := p.EvictPage()
		if err != nil {
			return hit, err
		}

		// Reset the new Frame, Add all the details we need
		p.Frames[NewFrameId].Reset()
		p.Frames[NewFrameId].PageOffsetInFile = PageStartAddress
		p.Frames[NewFrameId].ReferenceCount = 1 // this is the first read, we assume it's being referenced at least once
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
	// check if the key exists
	hit := CheckMapKeyPresent(&p.Data, PageStartAddress)
	if hit {
		// if it does
		// modify the page
		TargetPageId := p.Data[PageStartAddress]
		p.Frames[TargetPageId].data = *Content // TODO should this be a os.copy instead of an assignment?
		// mark as dirty
		p.Frames[TargetPageId].Dirty = true
		// return success
		return true, nil
	} else {
		// if it doesn't
		// evict a frame
		NewFrameId, err := p.EvictPage()
		if err != nil {
			return hit, err
		}

		// Reset the new Frame, Add all the details we need
		p.Frames[NewFrameId].Reset()
		p.Frames[NewFrameId].PageOffsetInFile = PageStartAddress
		p.Frames[NewFrameId].ReferenceCount = 1 // this is the first read, we assume it's being referenced at least once
		// replace the frame's content with the written buffer
		p.Frames[NewFrameId].data = *Content
		// mark as dirty
		p.Frames[NewFrameId].Dirty = true
		// Write the page back to disk
		// NOTE modify this later - as of now, this is write-through
		_, err = p.FlushFrameToDisk(NewFrameId)
		if err != nil {
			return false, err
		}
		p.Frames[NewFrameId].Dirty = false
		// return a success
		return true, nil
	}
}

// Dereference a page
func (p *PageCache) DereferencePage(PageStartAddress int) error {
	hit := CheckMapKeyPresent(&p.Data, PageStartAddress)
	if hit {
		FrameId := p.Data[PageStartAddress]
		p.Frames[FrameId].ReferenceCount -= 1 // NOTE this should be an atomic operation, with some sort of safeguard to ensure that only processes that referenced the page can dereference it, this is a BUG Until then
		return nil
	} else {
		return fmt.Errorf("Cannot Dereference a Page that doesn't exist (Offset %d)!", PageStartAddress)
	}
}

// Function to flush all dirty pages in the cache to disk
func (p *PageCache) FlushCacheToDisk() error {
	for index, frame := range p.Frames {
		if frame.Dirty {
			_, err := p.FlushFrameToDisk(index)
			if err != nil {
				return err
			}
			p.Frames[index].Dirty = false
		}
	}
	return nil
}

// A function that writes a page frame from the cache to the disk
func (p *PageCache) FlushFrameToDisk(FrameID int) (bool, error) {
	// Write the page back to the file at the start offset
	written_count, err := p.file.WriteAt(
		p.Frames[FrameID].data,
		int64(p.Frames[FrameID].PageOffsetInFile),
	)
	if written_count != p.PageSize {
		return false, fmt.Errorf(
			"Could not Write a Full Page -> Expected %d Copied %d",
			p.PageSize,
			written_count,
		)
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// function to evict page and return Frame Index to evict
// For now, this is a simple brute-force LRU, will optimize later with CLOCK
func (p *PageCache) EvictPage() (int, error) {

	// find the first empty slot that isn't referenced
	min_ref_count := 10000 // TODO make this a reasonable max value
	min_ref_page_id := 0
	for i := 0; i < p.FrameCount; i++ {
		curr_frame := p.Frames[i]
		if curr_frame.ReferenceCount < min_ref_count {
			min_ref_count = curr_frame.ReferenceCount
			min_ref_page_id = i
		}
	}

	// Flush the target page to disk
	if p.Frames[min_ref_page_id].Dirty {
		_, err := p.FlushFrameToDisk(min_ref_page_id)
		if err != nil {
			return min_ref_page_id, err
		}
	}
	// Reset the Frame and delete the page entry in the hashmap
	p.Frames[min_ref_page_id].Reset()
	// Delete the mapping in the address translator
	delete(p.Data, min_ref_page_id)

	return min_ref_page_id, nil
}
