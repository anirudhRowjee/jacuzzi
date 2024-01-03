package jacuzzi

import (
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
