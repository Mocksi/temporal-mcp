package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
)

// hashWorkflowArgs produces a short (suitable for inclusion in workflow id) hash of the given arguments. Args must be
// json.Marshal-able.
func hashWorkflowArgs(allParams map[string]string, paramsToHash ...any) (string, error) {
	if len(paramsToHash) == 0 {
		log.Printf("Warning: No hash arguments provided - will hash all arguments. Please replace {{ hash }} with {{ hash . }} in the workflowIDRecipe")
		paramsToHash = []any{allParams}
	}

	hasher := fnv.New32()
	for _, arg := range paramsToHash {
		// important: json.Marshal sorts map keys
		bytes, err := json.Marshal(arg)
		if err != nil {
			return "", err
		}
		_, _ = hasher.Write(bytes)
	}
	return fmt.Sprintf("%d", hasher.Sum32()), nil
}
