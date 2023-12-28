package jacuzzi

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

// map of the start offset to the page Frame itself Maybe storing it all in a map isn't the best idea? Feel like I should decouple the address mapping and the Frame access. Frames never get destructed and can be amortized on startup in terms of alloc cost, but these indices constantly change
type PageCache struct {
	PageSize  int
	Data      map[int]int // Data is the mapping of address to frames
	Frames    []Frame     // Frames is where all page frames live, never get deleted
	SlotCount int
}

// Start the pagecache
func (p *PageCache) Init(PageSize, slots int) {
	p.PageSize = PageSize
	p.SlotCount = slots
	p.Data = make(map[int]int, p.SlotCount)
	p.Frames = make([]Frame, p.SlotCount)
	// Initialize all allocs for frame
	for idx := range p.Frames {
		p.Frames[idx].Initialize(p, idx)
	}
}

// A function that reads a page into a buffer
// return signature (bool, error) ->
// ----> bool: True if Cache Hit, False if Cache Miss
func (p *PageCache) ReadPage(PageStartAddress int, Destination []byte) (bool, error) {
	return false, nil
}

// A function that writes a page from the application to the pagecache
// return signature (bool, error) ->
// ----> bool: True if Cache Hit, False if Cache Miss
func (p *PageCache) WritePage(PageStartAddress int, Content *[]byte) (bool, error) {
	return false, nil
}
