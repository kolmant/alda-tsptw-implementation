package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"alda-tsptw-implementation/entities"
	"alda-tsptw-implementation/utils"

	"github.com/yourbasic/bit"
)

type TimeWindow struct {
	Start float32
	End   float32
}

type Solution struct {
	Graph      entities.Graph
	TimeWindow []TimeWindow
	N          int
	CaseName   string
	Before     []map[int]bool
}

func main() {
	if len(os.Args) < 3 {
		fmt.Printf("Missing parameters.\n\nusage: go run main.go <file name> <timeout (ms)>\n\n")
		os.Exit(1)
	}
	fileName := os.Args[1]
	timeLimit, _ := strconv.Atoi(os.Args[2])

	r, err := os.Open("testdata/" + fileName)
	if err != nil {
		fmt.Println("Error reading file:", err)
		fmt.Printf("Tip: <fileName> must not contain its path.\n\n")
		os.Exit(2)
	}

	var i, j, n int
	var start, end float32
	fmt.Fscanf(r, "%d", &n)

	distances := make([][]float32, n)
	timeWindows := make([]TimeWindow, n)
	before := make([]map[int]bool, n)

	for i = 0; i < n; i++ {
		distances[i] = make([]float32, n)
		before[i] = map[int]bool{}
		for j = 0; j < n; j++ {
			fmt.Fscanf(r, "%f", &distances[i][j])
		}
	}

	graph := entities.Graph{TravelTime: distances, N: n}

	for i = 0; i < n; i++ {
		fmt.Fscanf(r, "%f %f\n", &start, &end)
		timeWindows[i] = TimeWindow{Start: start, End: end}
	}

	solution := Solution{Graph: graph, TimeWindow: timeWindows, N: n, CaseName: fileName, Before: before}

	_, _ = executeWithTimeout(context.Background(), "solve", func() {
		solution.Solve()
	}, time.Duration(timeLimit)*time.Millisecond)
}

func (s *Solution) Solve() {
	var k int
	var j int

	for j = 0; j < s.N; j++ {
		timeWindowJ := s.TimeWindow[j]

		//Test #2: BEFORE function implementation
		for k = 0; k < s.N; k++ {
			if j != k {
				timeWindowK := s.TimeWindow[k]
				distance := s.Graph.GetDistance(k, j)
				if distance != entities.PRUNED && timeWindowK.Start+distance > timeWindowJ.End {
					s.Before[k][j] = true
				}
			}
		}
	}

	// Initialize ξ1={({1},1,0)} and F({1},1,0)= 0
	epsilonKminusone := make([]entities.Epsilon, 0)
	initialSet := entities.Set{Mapa: new(bit.Set).Add(0), N: s.N, Elems: 1, Camino: "0 "}
	epsilonKminusone = append(epsilonKminusone, entities.Epsilon{Set: &initialSet, I: 0, T: 0})

	epsilonsK := make([]entities.Epsilon, 0)

	F := entities.FType{Mapa: map[string]float32{}, N: s.N, Result: map[string][]entities.TType{}}
	F.Set(epsilonKminusone[0].Set, 0, 0, 0)

	// for(k=2,3,…..n) do
	for k = 1; k < s.N; k++ {
		epslionsMap := map[string]entities.Epsilon{}
		epsilonsK = make([]entities.Epsilon, 0, s.N)

		// for (𝑆, 𝑖, 𝑡) ∈ ξk-1 do
		for epsK := range epsilonKminusone {
			epsilonK := epsilonKminusone[epsK]
			for j = 0; j < s.N; j++ {
				distance := s.Graph.GetDistance(epsilonK.I, j)
				if epsilonK.I != j && distance > entities.PRUNED {

					// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
					feasible := true

					// Test #2
					beforeJ := s.Before[j]
					for tmp := range beforeJ {
						if ok := epsilonK.Set.Mapa.Contains(tmp); !ok {
							feasible = false
							break
						}
					}

					if !feasible {
						continue
					}

					newEpsilon := epsilonK.Copy()
					Sprime, ok := newEpsilon.Set.Add(j)

					// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
					if ok {
						Tprime := utils.Max(s.TimeWindow[j].Start, epsilonK.T+distance)
						timeW := s.TimeWindow[j]

						// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
						if timeW.Start <= Tprime && Tprime <= timeW.End {
							fResult := F.Get(epsilonK.Set, epsilonK.I, epsilonK.T) + distance

							// Modified dominance test
							feasible = true
							existentEps, ok := epslionsMap[Sprime.GetStrRep()]
							if ok {
								if existentEps.T <= Tprime && F.Get(existentEps.Set, existentEps.I, existentEps.T) <= fResult {
									feasible = false

									break
								}
							}

							if feasible {
								feasibleEpsilon := entities.Epsilon{Set: &Sprime, I: j, T: Tprime}

								// update 𝐹(𝑆′,𝑗,𝑡′) = 𝐹(𝑆,𝑖,𝑡) + 𝑐𝑖𝑗 (𝑐𝑖𝑗 is already included in T𝑖𝑗)
								F.Set(&Sprime, j, Tprime, fResult)

								epsilonsK = append(epsilonsK, feasibleEpsilon)

								epslionsMap[Sprime.GetStrRep()] = feasibleEpsilon
							}
						}
					}
				}
			}
		}
		epsilonKminusone = epsilonsK
	}

	NMap := GenerateStrRepForN(s.N)

	var result float32 = 10000000000000
	var bestPath string

	for j = 0; j < s.N; j++ {
		distance := s.Graph.GetDistance(j, 0)
		if distance > entities.PRUNED {
			timeW := s.TimeWindow[j]
			for _, tmp := range F.Result[fmt.Sprintf("%s_%d", NMap, j)] {
				if timeW.Start <= tmp.T && tmp.T <= timeW.End && tmp.Value+distance < result {
					result = tmp.Value + distance
					bestPath = tmp.Camino
				}
			}
		}
	}

	// Generar tour en base al makespan
	path := strings.Split(bestPath, " ")
	var elapsed float32 = 0.0
	if len(path) > 0 {
		prev, _ := strconv.Atoi(path[0])
		for i := 1; i < len(path); i++ {
			curr, _ := strconv.Atoi(path[i])
			elapsed += s.Graph.GetDistance(prev, curr)
			prev = curr
		}
	}

	if elapsed != 0.0 {
		result = elapsed
	} else {
		// Solución no encontrada
		result = -1
	}

	fmt.Printf("instance: %s\n", s.CaseName)
	fmt.Printf("Cost: %.2f\n", result)
	fmt.Printf("Permutation: %s\n", bestPath)
}

func GenerateStrRepForN(N int) string {
	var sb strings.Builder
	var i int
	for i = 0; i < N; i++ {
		fmt.Fprintf(&sb, "%t_", true)
	}

	return sb.String()
}

func executeWithTimeout(ctx context.Context, name string, cb func(), timeout time.Duration) (interface{}, error) {
	var err error
	var resp interface{}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	res := make(chan answer, 1)
	c, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	go func() {
		cb()

		if c.Err() == nil {
			select {
			case res <- answer{"", nil}:
			default:
			}
		}
	}()

	select {
	case msg := <-res:
		resp = msg.payload
		err = msg.err
	case <-c.Done():
		err = fmt.Errorf("timeout executing %s procedure", name)
	}

	if err != nil {
		fmt.Println(err)
	}

	return resp, err
}

type answer struct {
	payload interface{}
	err     error
}
