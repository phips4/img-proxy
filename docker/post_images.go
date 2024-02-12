package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

const (
	numWorkers = 6
)

func main() {
	filePath := "urls.txt"
	url := "http://172.18.0.7:8080/image?url="

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	urls := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(urls, &wg)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		urls <- url + line
	}

	close(urls)
	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}

func worker(urls <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range urls {
		err := makeGetRequest(url)
		if err != nil {
			fmt.Println("Error making GET request:", err)
		}
	}
}

func makeGetRequest(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	fmt.Printf("successful GET request to %s\n", url)
	return nil
}
