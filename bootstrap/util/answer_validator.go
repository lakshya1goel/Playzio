package util

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

var (
	once         sync.Once
	wordSet      map[string]struct{}
	wordListPath = "wordlist.txt"
)

func loadWords() {
	once.Do(func() {
		wordSet = make(map[string]struct{})

		file, err := os.Open(wordListPath)
		if err != nil {
			panic("Failed to open wordlist.txt: " + err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			wordSet[strings.ToLower(word)] = struct{}{}
		}

		if err := scanner.Err(); err != nil {
			panic("Error reading wordlist.txt: " + err.Error())
		}
	})
}

func IsWordValid(word string) bool {
	loadWords()
	_, exists := wordSet[strings.ToLower(word)]
	return exists
}

func ContainsSubstring(word, substr string) bool {
	return strings.Contains(strings.ToLower(word), strings.ToLower(substr))
}
