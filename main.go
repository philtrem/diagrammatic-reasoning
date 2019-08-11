package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	// TODO: USE OUTPUTS AS INPUTS ?
	rand.Seed(time.Now().UnixNano())
	toPrint := make([]gridNode, 0)
	for i := 0; i < 1000; i++ {
		grid := newGridNode(6)
		for grid.introduceNewSequence() {
		}
		if grid.isValid() {
			toPrint = append(toPrint, grid)
		}
	}
	for _, item := range toPrint {
		item.print()
	}
	fmt.Println(len(toPrint))
}

const X = 0
const Y = 1
const XY = 2

const None int = 0
const Input int = 1
const Operator int = 2
const Output int = 3
const OutputInput int = 4 // acts as output and input

type node struct {
	label int // None, Input, Operator, Output, OutputInput
	orientation int // X, Y, XY
	shared bool
	knowable bool
	countX *int // for operators: how many in a row
	countY *int // for operators: how many in a row
	top *node
	right *node
	bottom *node
	left *node
}

type gridNode []rowNode
type rowNode []*node

func (g gridNode) print() {
	var separator string
	for range g {
		separator += "-"
	}
	for _, row := range g {
		line := ""
		for _, node := range row {
			switch node.label {
			case Input:
				line += "I"
			case Output:
				line += "O"
			case OutputInput:
				line += "X"
			case Operator:
				switch node.orientation {
				case X: line += "→"
				case Y: line += "↓"
				case XY: line += "+"//"🕂"
				}
			case None:
				line += " "
			}
		}
		fmt.Println(line)
	}
	fmt.Println(separator)
}

// return column with index i as row
func (g gridNode) columnToRow(i int) rowNode {
	n := len(g)
	out := make(rowNode, 0, n)
	for _, row := range g {
		for j := range row {
			if j == i {
				out = append(out, row[j])
			}
		}
	}
	return out
}

// checks if row can accept sequence
func (row rowNode) canAcceptSequence() bool {
	for i, node := range row {
		if node.label == None {
			for j := i+1; j < len(row) - 2; j++ {
				if row[j].label == Input || row[j].label == Output {
					break
				} else if row[j+1].label == None {
					return true
				}
			}
		}
	}
	return false
}

// return a slice containing 2 slices containing indices (as int)
// the first for available spots on the x-axis
// the second for available spots on the y-axis
func (g gridNode) availableRowsColumns() [][]int{
	out := make([][]int, 2)
	for i, row := range g {
		if row.canAcceptSequence() {
			out[0] = append(out[0], i)
		}
	}
	rotated := g.rotated()
	for i, row := range rotated {
		if row.canAcceptSequence() {
			out[1] = append(out[1], i)
		}
	}
	return out
}

// returns a slice containing indices of cells that can accept an input
func availableCellsInput(row rowNode) []int {
	out := make([]int, 0)
	n := len(row)
	for i := 0; i < n-1; i++ {
		if row[i].label != None {
			continue
		} else if row[i+1].label == Input || row[i+1].label == Output {
			continue
		}
		for j := i+2; j < n; j++ {
			if row[j].label == Input || row[j].label == Output {
				break
			} else if row[j].label == None {
				out = append(out, i)
				break
			}
		}
	}
	return out
}

// returns a slice containing indices of cells that are compatible
// with a specified index for an input
func compatibleCellsOutput(row rowNode, i int) []int {
	n := len(row)
	out := make([]int, 0)
	// assumes i is index of valid input,
	// so we don't have to check the first 2 indices,
	for j := i+2; j < n; j++ {
		if row[j].label == Input || row[j].label == Output {
			break
		} else if row[j].label == None {
			out = append(out, j)
		}
	}
	return out
}

// generate new sequence (input -> T * n -> output) to be inserted
func newSequence(row rowNode, orientation int) rowNode {
	out := make(rowNode, len(row))
	for i := range out {
		out[i] = &node{}
	}
	validInputs := availableCellsInput(row)
	randomIndex := rand.Intn(len(validInputs))
	i := validInputs[randomIndex]
	validOutputs := compatibleCellsOutput(row, i)
	randomIndex = rand.Intn(len(validOutputs))
	j := validOutputs[randomIndex]
	for before := 0; before < i; before++ {
		out[before] = row[before]
	}
	for after := j+1; after < len(row); after++ {
		out[after] = row[after]
	}
	//head
	out[i].label = Input
	out[i].orientation = orientation
	//tail
	out[j].label = Output
	out[j].orientation = Output
	// body

	// set-up operators connecting Input and Output
	// also: sets the count for number of operators
	// connecting an Input to its Output
	var count int
	for k := i+1; k < j; k++ {
		// check if node is already set as Operator
		// if so: change orientation value to XY
		if row[k].label == Operator {
			out[k].label = Operator
			out[k].orientation = XY
			switch orientation {
			case X:
				out[k].countX = &count
				out[k].countY = row[k].countY
			case Y:
				out[k].countY = &count
				out[k].countX = row[k].countX
			}
			count++
		} else {
			out[k].label = Operator
			out[k].orientation = orientation
			switch orientation {
			case X:
				out[k].countX = &count
			case Y:
				out[k].countY = &count
			}
			count++
		}
	}
	return out
}

func (g *gridNode) insertRow(row rowNode, i int) {
	for j := range (*g)[i] {
		(*g)[i][j] = row[j]
	}
}

func (g *gridNode) insertRowAsColumn(row rowNode, i int) {
	for j := range *g {
		(*g)[j][i] = row[j]
		(*g)[j][i] = row[j]
	}
}

// calls newSequence and inserts sequence into grid
// returns true on success, false if no more space to fit sequence
func (g *gridNode) introduceNewSequence() bool {
	available := g.availableRowsColumns()
	x := available[0]
	y := available[1]
	if len(x) == 0 && len(y) == 0 {
		return false
	}
	randomIndex := rand.Intn(len(x) + len(y))
	orientation := X
	// if i > len(x)-1: we're referring to array y
	if randomIndex > len(x) - 1 {
		randomIndex -= len(x) // correcting index for array y
		orientation = Y
	}
	var sequence rowNode
	switch orientation {
	case X:
		i := x[randomIndex]
		sequence = newSequence((*g)[i], orientation)
		g.insertRow(sequence, i)
	case Y:
		i := y[randomIndex]
		sequence = newSequence(g.columnToRow(i), orientation)
		g.insertRowAsColumn(sequence, i)
	}
	// replace consecutive inputs (on different axes)
	// with operators, and replace inputs following outputs
	// (on different axes) with operators
	// (outputs become 'outputinput', ie. they're acting as both
	// outputs and then inputs
	for i, row := range *g {
		for j, node := range row {
			if node.label == Input {
				switch node.orientation {
				case X:
					if j > 0 {
						prevNode := (*g)[i][j-1]
						nextNode := (*g)[i][j+1]
						if prevNode.label == Input && prevNode.orientation == Y {
							node.label = Operator
							node.countX = nextNode.countX
							*node.countX++
						} else if prevNode.label == Output {

						}
					}
				case Y:
					if i > 0 {
						prevNode := (*g)[i-1][j]
						nextNode := (*g)[i+1][j]
						if prevNode.label == Input && prevNode.orientation == X {
							node.label = Operator
							node.countY = nextNode.countY
							*node.countY++
						}

					}
				}
			}
		}
	}
	return true
}

// TODO: REMOVE ? not needed so far
// rotate 90 degrees clockwise  AFFECTS ORIGINAL gridNode
func (g *gridNode) rotate() {
	n := len(*g)
	result := newGridNode(n)
	for i, row := range *g {
		for j, node := range row {
			result[j][n-i-1] = node
		}
	}
	for i, row := range result {
		(*g)[i] = row
	}
}

// TODO: REMOVE ? not needed so far
// rotate 90 degrees clockwise  RETURNS NEW gridNode
func (g *gridNode) rotated() gridNode {
	n := len(*g)
	out := newGridNode(n)
	for i, row := range *g {
		for j, node := range row {
			out[j][n-i-1] = node
		}
	}
	return out
}

func newGridNode(n int) gridNode {
	// generate grid
	out := make(gridNode, n)
	for i := range out {
		out[i] = make([]*node, n)
	}
	for _, row := range out {
		for i := range row {
			row[i] = &node{}
		}
	}
	// set-up top,right,bottom,left pointers where applicable
	for i, row := range out {
		for j, node := range row {
			if i > 0 {
				node.top = out[i-1][j]
			}
			if j < n-1 {
				node.right = out[i][j+1]
			}
			if i < n-1 {
				node.bottom = out[i+1][j]
			}
			if j > 0 {
				node.left = out[i][j-1]
			}
		}
	}
	return out
}

func (g gridNode) isValid() bool {
	operators := make([]*node, 0)
	for _, row := range g {
		for _, node := range row {
			if node.label == Operator {
				operators = append(operators, node)
			}
		}
	}
	// mark shared nodes as knowable in a first pass
	for _, node := range operators {
		if node.orientation == XY {
			node.knowable = true
			// adjusting count for adjacent nodes
			*node.countX--
			*node.countY--
		}
	}
	// mark single nodes as knowable in a second pass
	// * skipping shared nodes
	for _, node := range operators {
		switch node.orientation {
		case X:
			if *node.countX == 1 {
				node.knowable = true
			}
		case Y:
			if *node.countY == 1 {
				node.knowable = true
			}
		}
	}
	for _, node := range operators {
		if node.knowable == false {
			return false
		}
	}
	return true
}

