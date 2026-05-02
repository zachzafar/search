package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type TermData struct {
	DocId     int64
	Occurence int64
}

type InvertedIndex struct {
	invertedIndex map[string][]*TermData
}

func main() {

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go: <json file name>")
	}

	fileName := os.Args[1]

	data, err := os.ReadFile(fileName)

	if err != nil {
		fmt.Printf("Error while reading file")
		return
	}

	var result []map[string]interface{}
	err = json.Unmarshal(data, &result)

	if err != nil {
		fmt.Printf("Error while extracting JSON, %v", err)
		return
	}

	for key, value := range result {
		fmt.Printf("\n Checking item %v, \n", key)
		printKeys((value))
	}

}

func printKeys(jsonObj map[string]interface{}) {
	for key, value := range jsonObj {
		fmt.Printf(" Key: %s", key)

		if nestedMap, ok := value.(map[string]interface{}); ok {
			printKeys(nestedMap)
		}

	}
}

func (ii *InvertedIndex) createIndex(jsonObj map[string]interface{}) {
	for key, value := range jsonObj {

		// Convert the key from string to int
		docId, err := strconv.ParseInt(key, 10, 64)
		if err != nil {
			// Key is not a valid integer - skip or handle error
			fmt.Printf("Invalid docId key: %s\n", key)
			continue
		}
		if isAnyInt(value) {
			if nestedMap, ok := value.(map[string]interface{}); ok {
				ii.traverseObj(nestedMap, docId)
			}
		}

	}
}

func primitiveToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		return strconv.FormatBool(val)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	default:
		return "" // Not a primitive we care about
	}
}

func (ii *InvertedIndex) addToIndex(str string, docId int64) {
	if termData, ok := ii.invertedIndex[str]; ok {
		for _, dataPoint := range termData {
			if dataPoint.DocId == docId {
				dataPoint.Occurence += 1
				return
			}
			// term not found in document create new term Data
			ii.invertedIndex[str] = append(termData, &TermData{
				DocId:     docId,
				Occurence: 1,
			})

		}
	} else {
		ii.invertedIndex[str] = []*TermData{
			{DocId: docId, Occurence: 1},
		}
	}
}

func (ii *InvertedIndex) traverseObj(obj map[string]interface{}, docId int64) {
	for _, value := range obj {
		if str := primitiveToString(value); str != "" {
			ii.addToIndex(str, docId)
		}

		if arr, ok := value.([]interface{}); ok {
			for _, val := range arr {
				if str := primitiveToString(val); str != "" {
					ii.addToIndex(str, docId)
				} else if nestedObj, ok := val.(map[string]interface{}); ok {
					ii.traverseObj(nestedObj, docId)
				}
			}
		}

		if nestedObj, ok := value.(map[string]interface{}); ok {
			ii.traverseObj(nestedObj, docId)
		}

	}
}

func isAnyInt(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64:
		return true
	default:
		return false
	}
}
