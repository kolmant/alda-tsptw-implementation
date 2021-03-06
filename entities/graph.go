package entities

const PRUNED = -10000

type Graph struct {
	N          int
	TravelTime [][]float32
}

func (g *Graph) GetDistance(i, j int) float32 {
	return g.TravelTime[i][j]
}
