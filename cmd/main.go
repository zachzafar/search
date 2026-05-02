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

type Search struct {
	invertedIndex map[string][]*TermData
	ngramIndex    map[string][]string
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

	// for index, value := range result {
	// 	fmt.Printf("\n Checking item %v, \n", index)
	// 	printKeys(value)
	// }

	invertedIndex := New()

	invertedIndex.CreateIndex(result)

	writeToFile("test-inverted-index.json", invertedIndex.invertedIndex)
	writeToFile("test-ngram.json", invertedIndex.ngramIndex)

}

// func printKeys(jsonObj map[string]interface{}) {
// 	for key, value := range jsonObj {
// 		fmt.Printf(" Key: %s", key)

// 		if nestedMap, ok := value.(map[string]interface{}); ok {
// 			printKeys(nestedMap)
// 		}

// 	}
// }

func New() *Search {
	return &Search{
		invertedIndex: make(map[string][]*TermData),
		ngramIndex:    make(map[string][]string),
	}
}

func (ii *Search) CreateIndex(jsonObj []map[string]interface{}) {
	for index, value := range jsonObj {

		ii.traverseObj(value, int64(index))

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

func (ii *Search) addToIndex(str string, docId int64) {

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

	ii.addToNgram(str)
}

func (ii *Search) traverseObj(obj map[string]interface{}, docId int64) {
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

func writeToFile(filename string, v any) error {
	jsonData, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return fmt.Errorf("failed to Marshal json: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)

	if err != nil {
		return fmt.Errorf("failed to write filr %w", err)
	}

	return nil
}

func (ii *Search) addToNgram(str string) {
	patterns := generateNGramPatterns(str, 3)

	for _, pattern := range patterns {
		ii.ngramIndex[pattern] = append(ii.ngramIndex[pattern], str)
	}
}

func generateNGramPatterns(word string, n int) []string {
	patterns := []string{}
	length := len(word)

	if length < n {
		return patterns
	}

	patterns = append(patterns, fmt.Sprintf("^%s", word[:n-1]))

	for i := 0; i <= length-n; i++ {
		if i == 0 {
			continue
		}
		patterns = append(patterns, word[i:i+n])
	}

	patterns = append(patterns, fmt.Sprintf("%s$", word[length-(n-1):]))

	return patterns
}
