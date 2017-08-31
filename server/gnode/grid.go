package gnode

import (
	"fmt"
	"sort"
)

type Grid struct {
	nbNode         int
	nbLineConnect  int
	nbCrossConnect int
	nodes          []map[int]byte
	log            bool
	ref            [][2]int
	Nodes          [][]int
}

func CreateGrid(nbNode int, nbLineConnect int, nbCrossConnect int, isLog bool) *Grid {
	grid := &Grid{}
	grid.nbNode = nbNode
	grid.nbLineConnect = nbLineConnect
	grid.nbCrossConnect = nbCrossConnect
	grid.log = isLog
	logf.info("Create %d nodes grid nbLibneConnect=%d nbCrossConnect=%d\n", grid.nbNode, grid.nbLineConnect, grid.nbCrossConnect)
	grid.computeNbConnect()
	if grid.nbLineConnect > grid.nbNode {
		grid.nbLineConnect = grid.nbNode
	}
	if grid.nbCrossConnect > grid.nbNode {
		grid.nbCrossConnect = grid.nbNode
	}
	grid.makeLinks()
	grid.createPublicArray()
	grid.display()
	grid.printf("Create grid done\n")
	return grid
}

func (g *Grid) computeNbConnect() {
	if g.nbLineConnect != 0 || g.nbCrossConnect != 0 {
		return
	}
	if g.nbNode == 1 {
		g.nbLineConnect = 0
		g.nbCrossConnect = 0
		return
	}
	g.makeRef()
	last := g.ref[0]
	for x, params := range g.ref {
		if params[0] != 0 {
			//fmt.Printf("x=%d y=%d\n", x, y)
			if x == g.nbNode {
				g.nbLineConnect = params[0]
				g.nbCrossConnect = params[1]
				//fmt.Printf("nbConnect: %d\n", g.nbConnect)
				return
			}
			if x > g.nbNode {
				break
			}
			last = params
		}
	}
	g.nbLineConnect = last[0]
	g.nbCrossConnect = last[1]
}

func (g *Grid) makeRef() {
	g.ref = make([][2]int, 260, 260)
	g.ref[2] = [2]int{1, 0}
	g.ref[3] = [2]int{2, 0}
	g.ref[4] = [2]int{3, 0}
	g.ref[5] = [2]int{4, 0}
	g.ref[6] = [2]int{5, 1}
	g.ref[10] = [2]int{9, 0}
	g.ref[16] = [2]int{15, 1}
	g.ref[20] = [2]int{19, 2}
	g.ref[30] = [2]int{4, 2}
	g.ref[38] = [2]int{5, 2}
	g.ref[46] = [2]int{6, 2}
	g.ref[102] = [2]int{7, 2}
	g.ref[118] = [2]int{7, 3}
	g.ref[184] = [2]int{7, 4}
	g.ref[248] = [2]int{8, 4}
	g.ref[256] = [2]int{8, 4}
}

func (g *Grid) makeArrays() {
	g.nodes = make([]map[int]byte, g.nbNode, g.nbNode)
	if g.nbNode == 1 {
		return
	}
	for i, _ := range g.nodes {
		nodeMap := make(map[int]byte)
		g.nodes[i] = nodeMap
	}
}

func (g *Grid) printf(format string, args ...interface{}) {
	if !g.log {
		return
	}
	fmt.Printf(format, args...)
}

func (g *Grid) makeLinks() {
	logf.info("Make grid with nbNode=%d and nbLineConnect=%d nbCrossConnect=%d\n", g.nbNode, g.nbLineConnect, g.nbCrossConnect)
	g.makeArrays()
	if g.nbNode == 1 {
		return
	}
	g.computeLink()
}

func (g *Grid) computeLink() bool {
	if g.nbNode == 1 {
		return true
	}
	fmt.Printf("connect %d\n", g.nbLineConnect)
	g.addLineConnects()
	g.addCrossConnects()
	return true
}

func (g *Grid) addLineConnects() {
	for n, _ := range g.nodes {
		g.addLineConnectForOneNode(n)
	}
}

func (g *Grid) addLineConnectForOneNode(node int) {
	for n := 0; n < g.nbLineConnect; n++ {
		target := (node + n + 1) % g.nbNode
		if target != node {
			g.addConnection(node, target)
		}
	}
}

func (g *Grid) addConnection(a int, b int) {
	nMap := g.nodes[a]
	nMap[b] = 1
	nMapr := g.nodes[b]
	nMapr[a] = 1
}

func (g *Grid) addCrossConnects() {
	for n, _ := range g.nodes {
		g.addCrossConnectForOneNode(n)
	}
}

func (g *Grid) addCrossConnectForOneNode(node int) {
	pow2 := 1
	for s := 0; s < g.nbCrossConnect; s++ {
		pow2 = pow2 * 2
		step := g.nbNode / pow2
		if step == 0 {
			step = 1
		}
		for i := 1; i <= g.nbCrossConnect; i++ {
			target := (node + step*i) % g.nbNode
			if target != node {
				g.addConnection(node, target)
			}
		}
	}
}

func (g *Grid) format(val int) string {
	if g.nbNode < 100 {
		if val >= 0 && val < 10 {
			return fmt.Sprintf("0%d", val)
		}
		return fmt.Sprintf("%d", val)
	}
	if val >= 0 && val < 10 {
		return fmt.Sprintf("00%d", val)
	} else if val >= 10 && val < 100 {
		return fmt.Sprintf("0%d", val)
	}
	return fmt.Sprintf("%d", val)
}

func (g *Grid) createPublicArray() {
	g.Nodes = make([][]int, g.nbNode, g.nbNode)
	for n, nMap := range g.nodes {
		array := make([]int, len(nMap), len(nMap))
		index := 0
		for key, _ := range nMap {
			array[index] = key
			index++
		}
		sort.Ints(array)
		g.Nodes[n] = array
	}
}

func (g *Grid) display() {
	for i, array := range g.Nodes {
		fmt.Printf("Node %s: ", g.format(i))
		for _, val := range array {
			fmt.Printf(" %s", g.format(val))
		}
		fmt.Printf("\n")
	}
}
