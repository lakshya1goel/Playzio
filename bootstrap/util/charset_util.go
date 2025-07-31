package util

import (
	"bufio"
	"math/rand"
	"os"
	"strings"
	"sync"
)

var (
	once         sync.Once
	wordSet      map[string]struct{}
	wordList     []string
	wordListPath = "wordlist.txt"
)

func loadWords() {
	once.Do(func() {
		wordSet = make(map[string]struct{})
		wordList = make([]string, 0)

		file, err := os.Open(wordListPath)
		if err != nil {
			panic("Failed to open wordlist.txt: " + err.Error())
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			word := strings.TrimSpace(scanner.Text())
			wordSet[strings.ToLower(word)] = struct{}{}
			wordList = append(wordList, word)
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

func GenerateRandomWord() string {
	loadWords()
	if len(wordList) < 4100 {
		panic("wordlist.txt does not have 4100 words")
	}
	word := wordList[4099]
	if len(word) < 2 {
		return word
	}
	minLen := 2
	maxLen := 3
	if len(word) < maxLen {
		maxLen = len(word)
	}
	length := rand.Intn(maxLen-minLen+1) + minLen
	if len(word) == length {
		return word
	}
	start := rand.Intn(len(word) - length + 1)
	return word[start : start+length]
}
