package main

//Documentation: http://cse.iitkgp.ac.in/~yeteshc/OR_final.pdf

import (
	"context"
	"encoding/json"
	"first-try-tsptw/entities"
	"first-try-tsptw/utils"
	"fmt"
	_ "net/http/pprof"
	"strings"
	"sync"
	"time"

	"github.com/yourbasic/bit"
)

type TimeWindows struct {
	Start float32
	End   float32
}

type Solution struct {
	Graph       entities.Graph
	TimeWindows []TimeWindows
	N           int
}

type Set struct {
	Mapa   *bit.Set
	strRep string
	N      int
	Elems  int
	Camino string
}

func (e *Set) Copy() Set {
	copiedMap := new(bit.Set)

	e.Mapa.Visit(func(n int) (skip bool) {
		copiedMap.Add(n)

		return false
	})

	return Set{
		Mapa:   copiedMap,
		N:      e.N,
		Elems:  e.Elems,
		Camino: e.Camino,
	}
}

type Epsilon struct {
	Set *Set
	I   int
	T   float32
}

func (e *Epsilon) Copy() Epsilon {
	set := e.Set.Copy()
	return Epsilon{
		Set: &set,
		I:   e.I,
		T:   e.T,
	}
}

type TType struct {
	T      float32
	Value  float32
	Camino string
}

type FType struct {
	mapa   map[string]float32
	N      int
	Result map[string][]TType
}

func (f *FType) Set(s *Set, i int, t float32, value float32) {
	f.mapa[fmt.Sprintf("%s_%d_%f", s.GetStrRep(), i, t)] = value
	if s.Elems == f.N {
		key := fmt.Sprintf("%s_%d", s.GetLargeStrRep(), i)
		res, ok := f.Result[key]
		if !ok {
			res = make([]TType, 0)
		}
		res = append(res, TType{T: t, Value: value, Camino: s.Camino})
		f.Result[key] = res
	}
}

func (f *FType) Get(s *Set, i int, t float32) float32 {
	return f.mapa[fmt.Sprintf("%s_%d_%f", s.GetStrRep(), i, t)]
}

func (s *Set) GetStrRep() string {
	if s.strRep == "" {
		s.GenerateStrRep()
	}

	return s.strRep
}
func (s *Set) GetLargeStrRep() string {
	var sb strings.Builder
	var i int
	for i = 0; i < s.N; i++ {
		fmt.Fprintf(&sb, "%t_", s.Mapa.Contains(i))
	}
	return sb.String()
}

func (s *Set) GenerateStrRep() {
	s.strRep = s.Mapa.String()
}

func (s Set) Remove(i int) (Set, bool) {
	ok := s.Mapa.Contains(i)
	if ok {
		s.Mapa.Delete(i)
		s.GenerateStrRep()
		s.Elems--
	}

	return s, ok
}

func (s Set) Add(i int) (Set, bool) {
	ok := s.Mapa.Contains(i)
	newSet := s.Mapa
	if !ok {
		newSet = s.Mapa.Add(i)
		s.GenerateStrRep()
		s.Elems++
		s.Camino = fmt.Sprintf("%s%d ", s.Camino, i)
	}
	s.Mapa = newSet

	return s, !ok
}

func main() {
	var i, j, n int
	var start, end float32
	fmt.Scanf("%d", &n)
	distances := make([][]float32, n)
	timeWindows := make([]TimeWindows, n)
	for i = 0; i < n; i++ {
		distances[i] = make([]float32, n)
		for j = 0; j < n; j++ {
			fmt.Scanf("%f", &distances[i][j])
		}
	}
	graph := entities.Graph{TravelTime: distances, N: n}

	for i = 0; i < n; i++ {
		fmt.Scanf("%f %f\n", &start, &end)
		timeWindows[i] = TimeWindows{Start: start, End: end}
	}

	solution := Solution{Graph: graph, TimeWindows: timeWindows, N: n}

	/**
	for i = 0; i < n; i++ {
		fmt.Printf("%f %f\n", timeWindows[i].Start, timeWindows[i].End)
	}
	**/

	wb := sync.WaitGroup{}
	for i = 0; i < n; i++ {
		wb.Add(1)
		solution.Prune(i, &wb)
	}
	wb.Wait()
	_, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, _ = executeWithTimeout(context.Background(), "solve", func() {
		solution.Solve(n > 5)
	}, 20*time.Minute)
}

func (s *Solution) Prune(i int, wb *sync.WaitGroup) {
	var j int
	for j = 0; j < s.N; j++ {
		travelTimeI := s.TimeWindows[i]
		travelTimeJ := s.TimeWindows[j]
		if s.Graph.GetDistance(i, j)+travelTimeI.Start > travelTimeJ.End {
			//s.Graph.Prune(i, j)
			//fmt.Printf("Arc %d %d PRUNEDDD\n", i, j)
		}
	}
	wb.Done()
}

func (s *Solution) Solve(feasibleTestsEnabled bool) {
	start := time.Now()
	var k int
	var j int

	before := make([]map[int]bool, s.N)

	for j = 0; j < s.N; j++ {
		before[j] = map[int]bool{}
	}
	if feasibleTestsEnabled {
		for j = 0; j < s.N; j++ {
			timeWindowJ := s.TimeWindows[j]

			//Test #2: BEFORE function implementation
			for k = 0; k < s.N; k++ {
				if j != k {
					timeWindowK := s.TimeWindows[k]
					distance := s.Graph.GetDistance(k, j)
					if distance != entities.PRUNED && timeWindowK.Start+distance > timeWindowJ.End {
						before[k][j] = true
					}
				}
			}
		}
	}

	// Initialize ξ1={({1},1,0)} and F({1},1,0)= 0
	epsilonKminusone := make([]Epsilon, 0)
	initialSet := Set{Mapa: new(bit.Set).Add(0), N: s.N, Elems: 1, Camino: ""}
	epsilonKminusone = append(epsilonKminusone, Epsilon{Set: &initialSet, I: 0, T: 0})

	epsilonsK := make([]Epsilon, 0)

	F := FType{mapa: map[string]float32{}, N: s.N, Result: map[string][]TType{}}
	F.Set(epsilonKminusone[0].Set, 0, 0, 0)

	notFeasible := 0

	fmt.Println(time.Since(start).Milliseconds(), "ms  --  epslions[ 0 ]: ", len(epsilonKminusone), "no feasible sols:", notFeasible)

	// for(k=2,3,…..n) do
	for k = 1; k < s.N; k++ {
		epslionsMap := map[string]Epsilon{}
		_ = map[string]Epsilon{}
		epsilonsK = make([]Epsilon, 0, s.N)
		// for (𝑆, 𝑖, 𝑡) ∈ ξk-1 do
		for epsK := range epsilonKminusone {
			epsilonK := epsilonKminusone[epsK]
			for j = 0; j < s.N; j++ {
				distance := s.Graph.GetDistance(epsilonK.I, j)
				if epsilonK.I != j && distance > entities.PRUNED {
					// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
					feasible := true

					if feasibleTestsEnabled {
						// Test #2
						beforeJ := before[j]
						for tmp := range beforeJ {
							if ok := epsilonK.Set.Mapa.Contains(tmp); !ok {
								feasible = false
								break
							}
						}

						if !feasible {
							notFeasible++
							continue
						}
					}

					//fmt.Println("antes  ", epsilonK.Set.GetStrRep(), epsilonK.I, j)

					epsilonCopy := epsilonK.Copy()
					Sprime, ok := epsilonCopy.Set.Add(j)

					// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
					if ok {
						//fmt.Println("despues", epsilonK.Set.GetStrRep())
						//fmt.Println("prime  ", Sprime.GetStrRep())

						Tprime := utils.Max(s.TimeWindows[j].Start, epsilonK.T+distance)

						timeW := s.TimeWindows[j]

						// add the state (𝑆′,𝑗,𝑡′) to ξk only if (𝑆′,𝑗,𝑡′) passes elimination tests
						if timeW.Start <= Tprime && Tprime <= timeW.End {
							fResult := F.Get(epsilonK.Set, epsilonK.I, epsilonK.T) + distance

							// Dominance test
							feasible := true
							if feasibleTestsEnabled {
								existentEps, ok := epslionsMap[Sprime.GetStrRep()]
								if ok {
									//log.Println("factible? ", existentEps.T, Tprime, fResult)
									if existentEps.T <= Tprime && F.Get(existentEps.Set, existentEps.I, existentEps.T) <= fResult {
										feasible = false

										break
									}
								}
							}

							if feasible {
								feasibleEpsilon := Epsilon{Set: &Sprime, I: j, T: Tprime}

								// update 𝐹(𝑆′,𝑗,𝑡′) = 𝐹(𝑆,𝑖,𝑡) + 𝑐𝑖𝑗 (𝑐𝑖𝑗 is already included in T𝑖𝑗)
								F.Set(&Sprime, j, Tprime, fResult)
								//log.Println(Sprime, j, Tprime, fResult)

								epsilonsK = append(epsilonsK, feasibleEpsilon)

								epslionsMap[Sprime.GetStrRep()] = feasibleEpsilon
							}
						}
					}
				}
			}
		}
		fmt.Println("epslions[", k, "]: ", toString(epsilonsK))
		fmt.Println(time.Since(start).Milliseconds(), "ms  --  epslions[", k, "]: ", len(epsilonsK), "no feasible sols:", notFeasible)
		epsilonKminusone = epsilonsK
	}

	fmt.Println("Resultado: ", toString(F.Result))

	NMap := GenerateStrRepForN(s.N)
	var result float32 = 1000000
	var mejorCamino string

	for j = 0; j < s.N; j++ {
		distance := s.Graph.GetDistance(j, 0)
		if distance > entities.PRUNED {
			timeW := s.TimeWindows[j]
			for _, tmp := range F.Result[fmt.Sprintf("%s_%d", NMap, j)] {
				if timeW.Start <= tmp.T && tmp.T <= timeW.End && tmp.Value+distance < result {
					result = tmp.Value + distance
					mejorCamino = tmp.Camino
				}
			}
		}
	}

	fmt.Println(mejorCamino)
	fmt.Println(result)
	fmt.Printf("%.2f\n", result)
}

func GenerateStrRepForN(N int) string {
	var sb strings.Builder
	var i int
	for i = 0; i < N; i++ {
		fmt.Fprintf(&sb, "%t_", true)
	}

	return sb.String()
}

func toString(i interface{}) string {
	b, err := json.Marshal(&i)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return string(b)
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

	start := time.Now()
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
		err = fmt.Errorf("timeout executing %s", name)
	}

	fmt.Println("time elapsed:", time.Since(start).Milliseconds(), "ms.")
	return resp, err
}

type answer struct {
	payload interface{}
	err     error
}
