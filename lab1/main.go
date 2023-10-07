package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"
)

type Task struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Resources    int      `json:"resources"`
	Requirements []string `json:"requirements"`
	Duration     int      `json:"duration"`
}

type CombinationWithTime struct {
	Task     [][]Task
	Duration int
}

type Combinator struct {
	sourceChan    []chan [][]Task
	calculateChan chan [][]Task
	outChan       chan CombinationWithTime
	normilazeChan chan [][]Task
}

const NumOfWorkers int = 3

func MakeCombinator(numOfWorkers int) Combinator {
	c := Combinator{sourceChan: make([]chan [][]Task, numOfWorkers), calculateChan: make(chan [][]Task, 10000), outChan: make(chan CombinationWithTime), normilazeChan: make(chan [][]Task, 10000)}

	for i := 0; i < numOfWorkers; i += 1 {
		c.sourceChan[i] = make(chan [][]Task, 10)
	}

	return c
}

func PrintCombination(c [][]Task) {
	fmt.Print("///\nCombination:\n")
	for i := 0; i < len(c); i += 1 {
		fmt.Println(i, " worker queue: ", c[i])
	}
}

func ValidateQueue(arr []Task) bool {
	for i := 0; i < len(arr); i += 1 {
		if len(arr[i].Requirements) > 0 {
			for j := i; j < len(arr); j += 1 {
				if slices.IndexFunc(arr[i].Requirements, func(c string) bool { return c == arr[j].Id }) > -1 {
					return false
				}
			}
		}
	}

	return true
}

func ValidateCombination(c [][]Task) bool {
	for i := 0; i < len(c); i += 1 {
		if !ValidateQueue(c[i]) {
			return false
		}
	}

	return true
}

func CopyCombination(c [][]Task) [][]Task {
	var result [][]Task = make([][]Task, len(c))

	for i := 0; i < len(c); i += 1 {
		result[i] = make([]Task, len(c[i]))
		copy(result[i], c[i])
	}

	return result
}

func CalculateTimeOfChunk(c []Task, channel chan<- int, wg *sync.WaitGroup) {
	execTime := 0

	for i := 0; i < len(c); i += 1 {
		execTime += c[i].Duration
	}

	channel <- execTime
	wg.Done()
}

func CalculateExecTimeOfQueue(c []Task, channel chan<- int, wg *sync.WaitGroup) {
	execTime := 0
	sizeOfChunk := 2000
	numOfChunks := len(c) / sizeOfChunk
	if len(c)%sizeOfChunk > 0 {
		numOfChunks += 1
	}
	var wg_in sync.WaitGroup
	ch := make(chan int, numOfChunks)

	for i := 0; i < numOfChunks; i += 1 {
		wg_in.Add(1)
		if i+1 == numOfChunks {
			go CalculateTimeOfChunk(c[i*sizeOfChunk:], ch, &wg_in)
		} else {
			go CalculateTimeOfChunk(c[i*sizeOfChunk:(i+1)*sizeOfChunk], ch, &wg_in)
		}
	}

	wg_in.Wait()
	close(ch)

	for et := range ch {
		execTime += et
	}
	channel <- execTime
	wg.Done()
}

func CalculateExecTimeOfCombination(c [][]Task) int {
	maxTime := -1

	var wg sync.WaitGroup

	channel := make(chan int, len(c))

	for i := 0; i < len(c); i += 1 {
		wg.Add(1)
		go CalculateExecTimeOfQueue(c[i], channel, &wg)
	}

	wg.Wait()
	close(channel)

	for execTime := range channel {
		if maxTime == -1 || maxTime < execTime {
			maxTime = execTime
		}
	}

	return maxTime
}

func MoveTaskToQueue(c [][]Task, from int, to int, index int) [][]Task {
	newCombination := CopyCombination(c)
	task := newCombination[from][index]

	newCombination[to] = append(newCombination[to], task)
	if index+1 == len(newCombination[from]) {
		newCombination[from] = append(make([]Task, 0), newCombination[from][:index]...)
	} else {
		t := append(make([]Task, 0), newCombination[from][:index]...)
		newCombination[from] = append(t, newCombination[from][index+1:]...)
	}

	return newCombination
}

func WorkerMoves(c [][]Task, workerIndex int, size int, out []chan [][]Task) {
	if size == 0 {
		for i := 0; i < len(out); i += 1 {
			out[i] <- c
		}
		return
	}

	for i := 0; i < len(c[workerIndex]); i += 1 {
		newCombination := MoveTaskToQueue(c, workerIndex, workerIndex+1, i)
		if !ValidateCombination(newCombination) {
			continue
		}
		WorkerMoves(newCombination, workerIndex, size-1, out)
	}
}

func WorkerMovesWrapper(c [][]Task, workerIndex int, size int, out []chan [][]Task, group *sync.WaitGroup) {
	defer group.Done()
	WorkerMoves(c, workerIndex, size, out)
}

func ListenToGenerateCombinations(workerIndex int, in chan [][]Task, out []chan [][]Task, group *sync.WaitGroup) {
	defer group.Done()
	var innerGroup sync.WaitGroup
	for c := range in {
		for i := 1; i < len(c[workerIndex]); i += 1 {
			innerGroup.Add(1)
			go WorkerMovesWrapper(c, workerIndex, i, out, &innerGroup)
		}
	}
	innerGroup.Wait()
	close(out[0])
}

func NormalizeOrder(comb *[][]Task) *[][]Task {
	for i := 0; i < len(*comb); i += 1 {
		taskStartTime := 0
		for j := 0; j < len((*comb)[i]); j += 1 {
			if len((*comb)[i][j].Requirements) > 0 {
				for k := 0; k < len(*comb); k += 1 {
					otherTaskStartTime := 0
					if i == k {
						continue
					} else {
						for l := 0; l < len((*comb)[k]); l += 1 {
							if slices.IndexFunc((*comb)[i][j].Requirements, func(t string) bool { return t == (*comb)[k][l].Id }) > -1 {
								if otherTaskStartTime+(*comb)[k][l].Duration > taskStartTime {
									newComb := CopyCombination(*comb)
									newComb[i] = append(make([]Task, 0), (*comb)[i][:j]...)
									newComb[i] = append(newComb[i], Task{Id: "-1", Duration: otherTaskStartTime + (*comb)[k][l].Duration - taskStartTime, Name: "Idle", Requirements: []string{}, Resources: 0})
									newComb[i] = append(newComb[i], (*comb)[i][j:]...)
									return &newComb
								}
							}
							otherTaskStartTime += (*comb)[k][l].Duration
						}
					}
				}
			}
			taskStartTime += (*comb)[i][j].Duration
		}
	}

	return comb
}

func NormalizeResources(comb *[][]Task) []*[][]Task {
	result := []*[][]Task{}
	for i := 0; i < len(*comb); i += 1 {
		taskStartTime := 0
		for j := 0; j < len((*comb)[i]); j += 1 {
			taskEndTime := taskStartTime + (*comb)[i][j].Duration
			awaibleResources := NumOfWorkers - (*comb)[i][j].Resources
			if (*comb)[i][j].Resources > 0 {
				for k := 0; k < len(*comb); k += 1 {
					otherTaskStartTime := 0
					if i == k {
						continue
					} else {
						for l := 0; l < len((*comb)[k]) && otherTaskStartTime < taskEndTime; l += 1 {
							otherTaskEndTime := otherTaskStartTime + (*comb)[k][l].Duration
							if !(otherTaskStartTime < taskStartTime && otherTaskEndTime <= taskStartTime) && (*comb)[k][l].Resources > awaibleResources {
								firstCombination := CopyCombination(*comb)

								firstCombination[k] = append(make([]Task, 0), firstCombination[k][:l]...)
								firstCombination[k] = append(firstCombination[k], Task{Id: "-1", Name: "Idle", Duration: taskEndTime - otherTaskStartTime, Requirements: []string{}, Resources: 0})
								firstCombination[k] = append(firstCombination[k], (*comb)[k][l:]...)

								result = append(result, &firstCombination)

								secondCombination := CopyCombination(*comb)

								secondCombination[i] = append(make([]Task, 0), secondCombination[i][:j]...)
								secondCombination[i] = append(secondCombination[i], Task{Id: "-1", Name: "Idle", Duration: otherTaskEndTime - taskStartTime, Requirements: []string{}, Resources: 0})
								secondCombination[i] = append(secondCombination[i], (*comb)[i][j:]...)

								result = append(result, &secondCombination)

								return result
							}

							otherTaskStartTime = otherTaskEndTime
						}
					}
				}
			}
			taskStartTime = taskEndTime
		}
	}

	return result
}

func NormalizeFunc(comb *[][]Task, outChan chan [][]Task, group *sync.WaitGroup) {
	defer group.Done()

	innerComb := comb

	for {
		innerStruct := NormalizeOrder(innerComb)

		if innerStruct == innerComb {
			break
		}

		innerComb = innerStruct
	}

	for {
		newCombs := NormalizeResources(innerComb)

		for i := 0; i < len(newCombs); i += 1 {
			group.Add(1)
			go NormalizeFunc(newCombs[i], outChan, group)
		}

		if len(newCombs) > 0 {
			return
		} else {
			break
		}
	}

	outChan <- *innerComb
}

func ListenToNormalize(in chan [][]Task, out chan [][]Task, group *sync.WaitGroup) {
	defer group.Done()
	defer close(out)

	var innerGroup sync.WaitGroup

	for c := range in {
		innerGroup.Add(1)
		NormalizeFunc(&c, out, &innerGroup)
	}

	innerGroup.Wait()
}

func ListenToCalculateTime(in chan [][]Task, out chan CombinationWithTime, group *sync.WaitGroup) {
	defer group.Done()
	innerStruct := CombinationWithTime{Duration: -1}

	for c := range in {
		calculatedTime := CalculateExecTimeOfCombination(c)
		if innerStruct.Duration == -1 || innerStruct.Duration > calculatedTime {
			innerStruct = CombinationWithTime{Duration: calculatedTime, Task: c}
		}
	}
	out <- innerStruct
}

func ListenToSelectOptimal(in chan CombinationWithTime, group *sync.WaitGroup) {
	defer group.Done()
	var innerStruct CombinationWithTime = CombinationWithTime{Duration: -1}

	for c := range in {
		if innerStruct.Duration == -1 || innerStruct.Duration > c.Duration {
			innerStruct = c
		}
	}

	fmt.Println("Optimal time: ", innerStruct.Duration)
	PrintCombination(innerStruct.Task)
}

func (c *Combinator) Run(startCombination [][]Task) {
	var wg sync.WaitGroup

	for i := 0; i < len(c.sourceChan); i += 1 {
		wg.Add(1)
		if i+1 == len(c.sourceChan) {
			go ListenToGenerateCombinations(i, c.sourceChan[i], append(make([]chan [][]Task, 0), c.normilazeChan), &wg)
		} else {
			t := append(make([]chan [][]Task, 0), c.sourceChan[i+1:]...)
			t = append(t, c.normilazeChan)
			go ListenToGenerateCombinations(i, c.sourceChan[i], t, &wg)
		}
	}

	var calcWait sync.WaitGroup
	for i := 0; i < 20; i += 1 {
		calcWait.Add(1)
		go ListenToCalculateTime(c.calculateChan, c.outChan, &calcWait)
	}

	wg.Add(1)
	go ListenToNormalize(c.normilazeChan, c.calculateChan, &wg)

	wg.Add(1)
	go func() {
		calcWait.Wait()
		close(c.outChan)
		wg.Done()
	}()

	wg.Add(1)
	go ListenToSelectOptimal(c.outChan, &wg)

	c.sourceChan[0] <- startCombination
	c.calculateChan <- startCombination
	close(c.sourceChan[0])

	wg.Wait()
}

func main() {
	fileReader, err := os.Open("tasks.json")

	if err != nil {
		fmt.Println("Cannot open file")
	}
	defer fileReader.Close()

	decoder := json.NewDecoder(fileReader)

	decoder.Token()

	tasks := make([]Task, 0)

	for decoder.More() {
		var t Task
		err := decoder.Decode(&t)
		if err != nil {
			fmt.Println("Error on read")
			return
		}

		tasks = append(tasks, t)
	}

	c := MakeCombinator(NumOfWorkers - 1)

	firstCombination := make([][]Task, NumOfWorkers)

	firstCombination[0] = tasks
	for i := 1; i < NumOfWorkers; i += 1 {
		firstCombination[i] = make([]Task, 0)
	}

	start := time.Now()
	c.Run(firstCombination)
	end := time.Since(start)
	fmt.Println("\nexec time: ", end)
}
