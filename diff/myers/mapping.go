package myers

import "fmt"

// the myers diff is based on line mapping
// but if you do not care about line-wise change, only block change,
// you can group all blocks together to make them act like lines

// ComputeBlockMapping
// a block is represented in a string
// the function computes
// the result is a mapping from each index in new to its
// counterpart in old, -1 if new.
// the result is 0-based, which is a historical design.
// we may optimize to 1-based in the future.xs
func ComputeBlockMapping(oldBlocks []string, newBlocks []string) map[int]int {
	m := make(map[int]int, len(newBlocks))
	operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
		m[newLine] = oldLine
	}, nil)
	return m
}

func ComputeBlockMappingV2(oldBlocks []string, newBlocks []string) (newToOld map[int]int) {
	newToOld = make(map[int]int, len(newBlocks))
	// newToOldUpdate = make(map[int]int,len(newBlocks))
	operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) { // on same
		newToOld[newLine] = oldLine
		fmt.Printf("M: %d %d\n", oldLine+1, newLine+1)
	}, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int) { // on update
		// newToOldUpdate[]
		fmt.Printf("U: %d %d; %d %d\n", oldLineStart+1, oldLineEnd+1, newLineStart+1, newLineEnd+1)
	})
	return
}

// TODO: currently not used, maybe the only important thing is finding sames, not updates or deletions.
// deleted of old
// func ComputeMapping(oldBlocks []string, newBlocks []string) (sameNewToOld map[int]int, updatedNewToOld map[int]int, deletedOld []int) {
// 	sameNewToOld = make(map[int]int, len(newBlocks))
// 	operations := operationsComplex(oldBlocks, newBlocks, func(oldLine, newLine int) {
// 		sameNewToOld[newLine] = oldLine
// 	}, func(oldLineStart, oldLineEnd, newLineStart, newLineEnd int) {

// 	})
// 	for _, op := range operations {
// 		switch op.Kind {
// 		case Delete:
// 			for i := op.I1; i <= op.I2; i++ {
// 				deletedOld = append(deletedOld, i)
// 			}
// 		case Insert:

// 		}
// 	}
// 	return
// }
