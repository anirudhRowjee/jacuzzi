package jacuzzi

// This is an implementation of a Min-Heap for the LRU architecture

// Function to get the length of the left child of a binary heap node
func LeftChildIndex(current_index int) int {
	return 2*current_index + 1
}

func RightChildIndex(current_index int) int {
	return 2*current_index + 2
}

func ParentIndex(current_index int) int {
	return current_index / 2
}

func Swap(data []int, idx1, idx2 int) {
	temp := data[idx1]
	data[idx1] = data[idx2]
	data[idx2] = temp
}

// This is the object being stored in the heap array
type DataObject struct {
	ID      int
	Counter int
}

// This is the heap itself
type MinHeap struct {
	size     int
	heapsize int
	data     []int
}

// Function to initialize the heap
func (h *MinHeap) Init(size int) {
	h.size = size
	h.heapsize = 0
	h.data = make([]int, size)
}

func (h *MinHeap) PeekItem() int { return h.data[0] }

// heapify down from the element at position int
func (h *MinHeap) HeapifyUp(position int) {
	if position == 0 {
		return
	}
	current_position := position
	for {
		parent_idx := ParentIndex(current_position)
		if parent_idx >= 0 {
			if current_position == 0 {
				break
			}
			parent_value := h.data[parent_idx]
			current_element := h.data[current_position]

			// Heap Invariant -> Parent has to be greater than the left and right child (Min Heap)
			if parent_value > current_element {
				Swap(h.data, parent_idx, current_position)
				current_position = parent_idx
			} else {
				current_position = 0 // exit
			}
		}
	}

}

// heapify up from the element at position int
func (h *MinHeap) HeapifyDown(position int) {
	current_position := position
	smallest := current_position

	left_index := LeftChildIndex(current_position)
	right_index := RightChildIndex(current_position)

	// check left child
	if left_index < h.heapsize-1 && h.data[left_index] < h.data[smallest] {
		smallest = left_index
	}
	// check right child
	if right_index < h.heapsize-1 && h.data[right_index] < h.data[smallest] {
		smallest = right_index
	}

	if smallest != position { // if a swap happened
		Swap(h.data, position, smallest)
		h.HeapifyDown(smallest)
	}
	// if the root is greater than the left or right child, swap them
}

func (h *MinHeap) AddItem(item int) {
	h.data[h.heapsize] = item
	h.heapsize += 1
	h.HeapifyUp(h.heapsize - 1)
}

func (h *MinHeap) PopItem() int {
	output := h.PeekItem()
	// swap the top and bottom variables
	Swap(h.data, 0, h.heapsize-1)
	// heapify down from root
	h.HeapifyDown(0)
	h.heapsize -= 1
	return output
}

func (h *MinHeap) UpdateItem() {}
func (h *MinHeap) DeleteItem() {}
