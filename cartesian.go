// Copyright 2015 NF Design UG (haftungsbeschraenkt). All rights reserved.
// Use of this source code is governed by the Apache License v2.0
// which can be found in the LICENSE file.

package main

func removeDuplicatesFromArray(arr [][]int) [][]int {
	length := len(arr) - 1
	for i := 0; i < length; i++ {
		for j := i + 1; j <= length; j++ {
			if arr[i][0] == arr[j][0] && arr[i][1] == arr[j][1] {
				arr[j] = arr[length]
				arr = arr[0:length]
				length--
				j--
			}
		}
	}
	return arr
}

func cartesianSelfProduct(n int, a []int) [][]int {

	var cp [][]int
	var nindex []int
	var b []int

	nindex = make([]int, n)

	for nindex != nil {
		b = make([]int, n)
		for i, j := range nindex {
			b[i] = a[j]
		}

		for i := len(nindex) - 1; i >= 0; i-- {
			nindex[i]++
			if nindex[i] < len(a) {
				break
			}
			nindex[i] = 0
			if i <= 0 {
				nindex = nil
				break
			}
		}
		cp = append(cp, b)
	}
	return cp
}

func generateContinuousIntArray(max int) []int {
	var init []int

	if max <= 0 {
		return []int{}
	}

	for i := 0; i < max; i++ {
		init = append(init, i+1)
	}
	return init
}

func sortAndRemoveIdentical(cp [][]int) [][]int {
	var r [][]int
	for _, value := range cp {

		if value[0] == value[1] {
			//There should be little use in testing a server with itself
			continue
		}

		//Sort values to make duplicate removal possible
		if value[0] < value[1] {
			r = append(r, []int{value[0], value[1]})
		} else {
			r = append(r, []int{value[1], value[0]})
		}

	}
	return r

}

func decrementValuesByOne(cp [][]int) [][]int {
	var r [][]int
	for _, value := range cp {
		r = append(r, []int{value[0] - 1, value[1] - 1})
	}
	return r

}

// generateTestPairs generates test pairs.
// It uses a post-processed cartesian productof a set of
// continuous integers with itself to determine mail account
// pairings for testing. This tests every sensible combination
// of the servers given.
func generateTestPairs(amountofaccounts int) [][]int {

	const amountofpairmembers = 2
	var r [][]int

	cp := cartesianSelfProduct(amountofpairmembers, generateContinuousIntArray(amountofaccounts))
	r = sortAndRemoveIdentical(cp)
	r = removeDuplicatesFromArray(r)
	r = decrementValuesByOne(r)
	return r

}
