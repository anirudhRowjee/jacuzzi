package jacuzzi

import (
	"sort"
	"testing"
)

func TestInit(t *testing.T) {
	if (1 + 1) != 2 {
		t.Error("Something's catastrophically wrong")
	}
}

func TestInitHeap(t *testing.T) {
	testSize := 4
	heap := MinHeap{}
	heap.Init(testSize)

	if heap.size != testSize {
		t.Errorf(
			"Could not Initialize Heap with Required Size Expected %d Got %d",
			testSize,
			heap.size,
		)
	}
	if len(heap.data) != testSize {
		t.Errorf(
			"Could not Initialize Heap Array with Required Size Expected %d Got %d",
			testSize,
			len(heap.data),
		)
	}
}

func TestHeapifyOnLoad(t *testing.T) {
	testSize := 4
	heap := MinHeap{}
	heap.Init(testSize)

	data := [4]int{3, 2, 1, 6}
	for _, v := range data {
		heap.AddItem(v)
	}

	if heap.PeekItem() != 1 {
		t.Errorf("Heapify not working! heap top not min element.")
	}

}

func TestHeapPop_small(t *testing.T) {
	testSize := 4
	heap := MinHeap{}
	heap.Init(testSize)

	data := [4]int{3, 2, 1, 6}
	for _, v := range data {
		heap.AddItem(v)
	}

	item := heap.PopItem()
	if item != 1 {
		t.Errorf("Heap Extraction not working! heap top not min element.")
	}

	item = heap.PopItem()
	if item != 2 {
		t.Errorf("Heap Extraction not working! heap top not min element.")
	}

	item = heap.PopItem()
	if item != 3 {
		t.Errorf("Heap Extraction not working! heap top not min element.")
	}

	item = heap.PopItem()
	if item != 6 {
		t.Errorf("Heap Extraction not working! heap top not min element.")
	}

}

func TestHeapPop_large(t *testing.T) {
	testSize := 10
	heap := MinHeap{}
	heap.Init(testSize)

	data := []int{10, 8, 5, 2, 3, 1, 7, 4, 6, 9}
	for _, v := range data {
		heap.AddItem(v)
	}

	// Sort the input slice
	sort.Ints(data)

	for _, v := range data {
		item := heap.PopItem()
		if item != v {
			t.Errorf("Heap Extraction not working! heap top not min element.")
		}
	}
}

func TestHeapPopWithDuplicates(t *testing.T) {
	testSize := 10
	heap := MinHeap{}
	heap.Init(testSize)

	data := []int{10, 8, 2, 2, 3, 1, 7, 7, 6, 9}
	for _, v := range data {
		heap.AddItem(v)
	}

	// sort the input slice
	sort.Ints(data)

	for _, v := range data {
		item := heap.PopItem()
		if item != v {
			t.Errorf("Heap Extraction not working! heap top not min element.")
		}
	}

}
