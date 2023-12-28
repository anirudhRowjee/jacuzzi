package main

import (
	"fmt"
	"testing"

	"golang.org/x/exp/slices"
)

// Test Initialization of a structure
func Test_initialize_pagecache(t *testing.T) {

	pagecache := PageCache{}
	pagecache.Init(
		10,         // slots
		"LRU",      // policy
		"test.bin", // Filename
	)

	fmt.Println(pagecache.PageSize)
}

// Check if the slots are being initialized properly
func Test_initialize_slots(t *testing.T) {

	slot_count := 10

	pagecache := PageCache{}
	pagecache.Init(
		slot_count, // slots
		"LRU",      // policy
		"test.bin", // Filename
	)
}

// Test: Read a page back into the main application
func Test_read_page_into_pagecache(t *testing.T) {

	slot_count := 10
	pagecache := PageCache{}
	pagecache.Init(
		slot_count, // slots
		"LRU",      // policy
		"test.bin", // Filename
	)

	// Read one page into the pagecache
	buffer_first := make([]byte, pagecache.PageSize)
	buffer_second := make([]byte, pagecache.PageSize)

	// Read a page into memory first
	hit, err := pagecache.ReadPage(0, &buffer_first)
	if err != nil {
		t.Error("Cache Read Failed >", err)
	}
	fmt.Println("Cache Hit Status:", hit)

	// Read page at the same offset into memory
	hit2, err2 := pagecache.ReadPage(0, &buffer_second)
	if err2 != nil {
		t.Error("Cache Read Failed >", err2)
	}
	fmt.Println("Cache Hit Status:", hit2)

	// compare the slices to see if we've gotten the same data
	if slices.Compare(buffer_first, buffer_second) != 0 {
		t.Error("Could not establish coherence between buffers > ")
	}

}

// Test: Flush a page to a file, read it back to see if it's modified
func Test_flush_page_to_file(t *testing.T) {
	slot_count := 10

	// Initialize a PageCache
	pagecache := PageCache{}
	pagecache.Init(
		slot_count, "LRU", "test.bin",
	)

	// Write a page to the cache

	// Flush the page

	// read it back into the pagecache

	// verify that the persisted changes have actually been persisted

}

// Test: Write a page back into the file
func Test_write_page_to_pagecache(t *testing.T) {
	// If we test this through the write function itself, we create
	// a cyclic dependency on this test

	// Create a pagecache
	slot_count := 10
	pagecache := PageCache{}
	pagecache.Init(
		slot_count, // slots
		"LRU",      // policy
		"test.bin", // Filename
	)

	// Read a page into the cache
	in_buffer := make([]byte, pagecache.PageSize)
	pagecache.ReadPage(0, &in_buffer)

	// manually modify the page

	// flush the page to disk (this is what we're testing)
	// read the page back from disk
	// compare to see if the blocks are the same
}
