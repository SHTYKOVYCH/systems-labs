package main

import (
	"encoding/json"
	"fmt"
	"os"
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

func GenerateCombination(numOfTasks int) {
	c := make([][]Task, 0)

	for i := 0; i < 3; i += 1 {
		for j := 0; j < numOfTasks; j += 1 {
			c = append(c, make([]Task, 0))

			for k := 0; k < j/(i+1); k += 1 {
				c[0] = append(c[0], Task{Duration: 1})
			}
		}
	}

	fmt.Println("generated item")

	stringToWrite, _ := json.Marshal(c)

	os.WriteFile("test_combination.json", stringToWrite, 0777)
}

//func ReadOrCreateAndRead(numOfTasks int) [][]Task {
//	reader, err := os.Open("test_combination.json")
//
//	if err != nil {
//		GenerateCombination(numOfTasks)
//		return ReadOrCreateAndRead(numOfTasks)
//	}
//	defer reader.Close()
//
//	decoder := json.NewDecoder(reader)
//	
//	var combination [][]Task
//	
//	decoder.Token()
//
//	for decoder.More() {
//		var t [][]Task
//		err := decoder.Decode(&t)
//		if err != nil {
//			fmt.Println("Error on read")
//			panic(nil)
//		}
//
//		combination = append(combination, t)
//	}
//
////	_ = json.Unmarshal(file, &combination)
//
//	return combination
//}

func test(a int) *int {
	return &a
}

func main() {
//	c := ReadOrCreateAndRead(5000)
//
//	var wg sync.WaitGroup
//	in := make(chan [][]Task, 100)
//	out := make(chan CombinationWithTime)
//
//	go func() {
//		for _ = range out {
//		}
//	}()
//	fmt.Println("Starting figing")
//
//	go func() {
//		for i := 0; i < 1; i += 1 {
//			in <- c
//		}
//		close(in)
//	}()
//	start := time.Now()
//
//	for i := 0; i < 1; i += 1 {
//		wg.Add(1)
//		go ListenToCalculateTime(in, out, &wg)
//	}
//
//	wg.Wait()
//	fmt.Println("\nexec time: ", time.Since(start))
	var a int = 5
	
	fmt.Println(&a == test(a))
}
