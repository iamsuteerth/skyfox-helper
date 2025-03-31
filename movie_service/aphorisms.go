package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

func readAphorisms(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open aphorisms file: %w", err)
	}
	defer file.Close()

	var aphorisms []string
	scanner := bufio.NewScanner(file)

	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		if i := strings.Index(string(data), "%"); i >= 0 {
			return i + 1, data[0:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	})

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			aphorisms = append(aphorisms, text)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading aphorisms: %w", err)
	}

	return aphorisms, nil
}

func getRandomAphorism(aphorisms []string) string {
	if len(aphorisms) == 0 {
		return "No aphorisms available"
	}
	return aphorisms[rand.Intn(len(aphorisms))]
}

func formatAphorism(aphorism string) (string, string) {
	parts := strings.Split(aphorism, " - ")
	if len(parts) < 2 {
		return aphorism, ""
	}
	return parts[0], parts[1]
}
