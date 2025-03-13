package pkg

import (
	"fmt"
	"sync"
)

func BubbleSort(array []int, l int, r int, result chan []int) {
	t := []int{}
	if l >= r {
		result <- t
		return 
	}
	for i := r - 1; i > l; i-- {
		for j := l; j < i; j++ {
			if array[j] > array[j + 1] {
				array[j] = array[j + 1] + array[j]
				array[j + 1] = array[j] - array[j + 1]
				array[j] = array[j] - array[j + 1]
			}
		}
	}
	for i := l; i < r; i++ {
		t = append(t, array[i])
	}
	fmt.Printf("done\n")
	result <- t
	return	
}

func Sort(array []int) {
	result := make(chan []int)
	size := len(array)
	wg := sync.WaitGroup{}
	chunks := 0
	left := 0
	right := 0 
	chunkSize := size / 4 
	if size < 4 {
		chunkSize = size
		chunks = 1
	} else if size % 4 == 0 {
		chunks = 4
	} else if size % 4 > 0 {
		chunks = 5
	}

	
	for i := 0;i < chunks; i++ {
		wg.Add(1)
		right = left + chunkSize
		if right >= size {
			right = size
		}
		fmt.Printf("calling sort on left %d right %d \n", left, right)
		go BubbleSort(array, left, right, result)
		left = right
	}

	
	for _, item := range array {
		fmt.Println(item)
	}
	
	x := [][]int{}
	for i := 0; i < chunks; i++ {
		res := <- result
		x = append(x, res)
	}

	pointers := make([]int, chunks)

	sizes := make([]int, chunks)

	for index, t := range x {
		sizes[index] = len(t)
	}

	for i := 0; i < size; i++ {
		mn := int(1e9)
		mni := int(-1)
		for index, pointer := range pointers {
			if pointer < sizes[index] {
				if mn > x[index][pointer] {
					mn = x[index][pointer]
					mni = index
				}
			}
		}
		if mni != -1 {
			pointers[mni] += 1
			array[i] = mn
		}
	}


	for _, item := range array {
		fmt.Println(item)
	}

}


func MainSort() {
	var n int
	fmt.Printf("Enter the size of array")
	fmt.Scan(&n)
	array := []int{}
	for i := 0; i < n; i++ {
		var x int
		fmt.Scan(&x)
		array = append(array, x)
	}
	Sort(array)
}

