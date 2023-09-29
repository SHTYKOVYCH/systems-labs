package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Task struct {
	Id           string   `json:"id"`
	Name         string   `json:"name"`
	Requirements []string `json:"requirements"`
	Resources    int      `json:"resources"`
	Duration     int      `json:"duration"`
}

func main() {
	file, err := os.OpenFile("tasks.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Print("Cannot open file\n")
		return
	}

	var numOfResources = 3
	var numOfTasks int = 3// rand.Intn(8) + 1
	var numOfTasksWithDependencies int = 1 //(rand.Intn(numOfTasks) + 1) % numOfTasks

	fmt.Println(numOfTasks, numOfTasksWithDependencies)

	var arr = make([]Task, 0, numOfTasks)

	for i := 0; i < numOfTasks-numOfTasksWithDependencies; i += 1 {
		arr = append(arr, Task{Id: strconv.Itoa(i + 1), Name: strings.Join([]string{"Task ", strconv.Itoa(i + 1)}, ""), Requirements: []string{}, Resources: rand.Intn(numOfResources) + 1, Duration: rand.Intn(100)})
	}

	for i := numOfTasks - numOfTasksWithDependencies; i < numOfTasks; i += 1 {
		deps := make([]string, 0, rand.Intn(numOfTasks-numOfTasksWithDependencies)+1)

		for j := 0; j < cap(deps); j += 1 {
			for {
				randIndex := rand.Intn(i) + 1
				if slices.Index(deps, strconv.Itoa(randIndex)) > -1 {
					continue
				}
				deps = append(deps, strconv.Itoa(randIndex))
				break
			}
		}
		arr = append(arr, Task{Id: strconv.Itoa(i + 1), Name: strings.Join([]string{"Task ", strconv.Itoa(i + 1)}, ""), Requirements: deps, Resources: rand.Intn(numOfResources) + 1, Duration: rand.Intn(100)})
	}

	stringToWrite, _ := json.Marshal(arr)

	file.Write(stringToWrite)
}
