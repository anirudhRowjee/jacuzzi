package main

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

// Initialize Logger

// First goal:
// Simple Cache that can read pages in and out, no replacement policy
// The precondition is that all reads and writes will be only of the fixed page size

// struct to describe a cache entry in the pagecache
type PageCacheSlot struct {
	mu           sync.Mutex // to ensure no use-after-evicts, easy concurrent writes, and so on
	data         []byte     // the raw data, one page size
	dirty        bool       // to indicate if this cache slot is dirty and needs a write back
	pinned       bool       // to indicate if this slot is pinned, which means we won't be removing it from the cache
	start_offset int        // the offset in the file that this slot entry's data belongs to
	size         int        // the page size; maybe we can push this up
	active       bool       // Is this slot actively used right now?
}

// function to initialize a pagecache slot
func (s *PageCacheSlot) Init(PageSize int) {
	s.data = make([]byte, PageSize)
	s.dirty = false
	s.active = false
	s.pinned = false
	s.start_offset = -1
	s.size = PageSize
}

// Reset the slot so we can reuse it
// Assumes we do not change the page size on runtime
func (s *PageCacheSlot) Reset() {
	s.dirty = false
	s.active = false
	s.pinned = false
	s.start_offset = -1
}

type PageCache struct {
	SlotsCount int                    // The number of cache slots in the pagecache; This cannot be changed at runtime
	PageSize   int                    // page size in bytes
	Policy     string                 // placeholder for policy
	Filename   string                 // File that this pagecache belongs to
	file       *os.File               // the actual file that this cache will be operating on
	slots      map[int]*PageCacheSlot // the slots

	// TODO Add MinHeap
	// TODO Add policy handler
	// TODO Add place to pin entries, add ddirty bit
	// TODO Interfaces
	//   - how to access data
	//   - how to flush data
	//   - how to pin pages
}

// Flush an offset to Disk; returns an error if the flush was unsuccessful
// As of right now, we are making two assumptions:
// 1. The page entry being flushed exists
// 2. The Flush is called outside of the write function (this is to promote extensibility -> We can make it a write-through or a write-back cache later based on what we want)

func (p *PageCache) FlushToDisk(offset int) error {
	slog.Info("Attempting to Flush Offset to File ",
		"offset", offset,
		"filename", p.Filename)

	buffer := make([]byte, p.PageSize)

	// read the slot
	p.slots[offset].mu.Lock()
	copied_bytes_count := copy(buffer, p.slots[offset].data)
	p.slots[offset].mu.Unlock()

	if copied_bytes_count != p.PageSize {
		return fmt.Errorf("Could not Copy all bytes from page!")
	}

	// write the content of the buffer onto the disk

	//
	return nil
}

// TODO should we return an error?
func (p *PageCache) Init(slots int, policy, filename string) error {

	slog.Info("Initializing PageCache", "filename", filename)
	// Basic initialization
	p.SlotsCount = slots
	p.PageSize = os.Getpagesize()
	p.Policy = policy
	p.Filename = filename

	// Open the underlying file that we're caching
	fd, err := os.OpenFile(p.Filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		// TODO see if we can error handle this better
		slog.Error("Couldn't open Cache File", "filename", p.Filename)
		// panic("Could Not Open Cache File!")
		return err
	}
	p.file = fd

	// Initialize the map
	p.slots = make(map[int]*PageCacheSlot)

	return nil

}

//	func (p *PageCache) SearchSlot(StartingOffset int) (PageCacheSlot, error) {
//		// TODO figure out some way to prevent copying here
//		// Copying is bad for memory usage, but much safer, because you don't risk the entry being invalidated while
//		// another goroutine has a pointer to it
//		// FIX solution: Mutex per cache slot? doesn't seem like too bad an idea
//		nil
//	}

// function to read pages from the cache
// Read file chunk [offset, offset+PageSize] into cache : returns True on Cache Hit, False on Cache Miss
func (p *PageCache) ReadPage(start_offset int, recipient *[]byte) (bool, error) {

	// Reason for not returning a pointer: there's a possible race condition where
	// the page being read gets evicted when there's still a pointer to that page in the code
	// somewhere else, so it's best to just copy it off
	// So what we're doing is we're accepting an address of a slice of bytes, so we can

	// here is also where we check if the page is already in the pagecache, then we decide to
	// read through or just return a cached bit of data

	// check if the key is there in the map
	present := false

	for k := range p.slots {
		if k == start_offset {
			// Found, Copy the page into the target buffer from the pagecache
			slog.Info("Cache Hit!",
				"offset", k,
			)
			present = true
			bytes_copied := copy(*recipient, p.slots[k].data)
			if bytes_copied != p.PageSize {
				// Malformed page?
				slog.Error("Malformed Page!",
					"size", p.PageSize,
					"bytes", bytes_copied,
				)
				return false, fmt.Errorf(
					"Malformed Data: Expected size %d Copied size %d\n",
					p.PageSize,
					bytes_copied,
				)
			}
			return present, nil
		}
	}

	// The page isn't present, we need to pull it into the cache
	if present == false {
		// NOTE this is Demand Paging!
		slog.Info("Cache Miss!",
			"offset", start_offset,
		)
		// for now, evict the 0th slot again
		replacement_slot_index := 0 // this is the index the new page entry will sit at
		// BUG this assumes that the slot exists at that index, This will do for now but not for later
		// BUG for now, replace the slotentry with a new slotentry

		new_slot := PageCacheSlot{}
		new_slot.Init(p.PageSize)

		// read from file into slot
		// Check if we can reuse the old slot after resetting it, otherwise make a new slot

		// TODO refactor t.Errorf usage here to use Slog
		bytes_read, err := p.file.Read(new_slot.data)
		if err != nil {
			return false, err
		}
		if bytes_read != p.PageSize {
			return false, fmt.Errorf(
				"Malformed Data from File: Expected size %d Copied size %d\n",
				p.PageSize,
				bytes_read,
			)
		}

		p.slots[replacement_slot_index] = &new_slot

		bytes_copied := copy(*recipient, p.slots[replacement_slot_index].data)
		if bytes_copied != p.PageSize {
			// Malformed page?
			return false, fmt.Errorf(
				"Malformed Data: Expected size %d Copied size %d\n",
				p.PageSize,
				bytes_copied,
			)
		}
		return false, nil
	}

	return present, nil
}

func (p *PageCache) WritePage(start_offset int, page_data *[]byte) (bool, error) {
	// here we can safely accept a pointer, as we don't have to worry about concurrent access
	// on the page being written back to the cache: This is a guarantee that has to be enforced by
	// the application code
	hit := false

	// check the offset
	for k := range p.slots {
		if k == start_offset {
			hit = true
			p.slots[start_offset].data = *page_data
			p.slots[start_offset].dirty = true
			break
		}
	}

	// if we are here and hit is false, we need to write the new cache entry ourselves
	if !hit {
		entry := PageCacheSlot{}
		entry.Init(p.PageSize)
		p.slots[start_offset] = &entry
	}

	return hit, nil
}

func main() {
	fmt.Println("Hello, from the pagecache!")
}
