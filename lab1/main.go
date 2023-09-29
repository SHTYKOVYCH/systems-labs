package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sync"
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
}

func MakeCombinator(numOfWorkers int) Combinator {
	c := Combinator{sourceChan: make([]chan [][]Task, numOfWorkers), calculateChan: make(chan [][]Task, 1000), outChan: make(chan CombinationWithTime, 1000)}

	for i := 0; i < numOfWorkers; i += 1 {
		c.sourceChan[i] = make(chan [][]Task)
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

	copy(result, c)

	for i := 0; i < len(c); i += 1 {
		copy(result[i], c[i])
	}

	return result
}

func CalculateExecTimeOfCombination(c [][]Task) int {
	maxTime := -1

	for i := 0; i < len(c); i += 1 {
		execTime := 0
		for j := 0; j < len(c[i]); j += 1 {
			execTime += c[i][j].Duration
		}

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

func ListenToCalculateTime(in chan [][]Task, out chan CombinationWithTime, group *sync.WaitGroup) {
	defer func() {
		group.Done()
		recover()
	}()
	innerStruct := CombinationWithTime{Duration: -1}

	for c := range in {
		calculatedTime := CalculateExecTimeOfCombination(c)
		if innerStruct.Duration == -1 || innerStruct.Duration > calculatedTime {
			innerStruct = CombinationWithTime{Duration: calculatedTime, Task: c}
		}
	}
	out <- innerStruct
	close(out)
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
			go ListenToGenerateCombinations(i, c.sourceChan[i], append(make([]chan [][]Task, 0), c.calculateChan), &wg)
		} else {
			t := append(make([]chan [][]Task, 0), c.sourceChan[i+1:]...)
			t = append(t, c.calculateChan)
			go ListenToGenerateCombinations(i, c.sourceChan[i], t, &wg)
		}
	}

	for i := 0; i < 2; i += 1 {
		wg.Add(1)
		go ListenToCalculateTime(c.calculateChan, c.outChan, &wg)
	}

	wg.Add(1)
	go ListenToSelectOptimal(c.outChan, &wg)

	c.sourceChan[0] <- startCombination
	c.calculateChan <- startCombination
	close(c.sourceChan[0])

	wg.Wait()
}

func main() {
	tasksUnparsed, _ := os.ReadFile("tasks.json")

	var tasks []Task

	_ = json.Unmarshal(tasksUnparsed, &tasks)

	const NumOfWorkers = 3

	c := MakeCombinator(NumOfWorkers)

	firstCombination := make([][]Task, NumOfWorkers)

	firstCombination[0] = tasks
	for i := 1; i < NumOfWorkers; i += 1 {
		firstCombination[i] = make([]Task, 0)
	}

	c.Run(firstCombination)
}
