/*
  - Copyright (c) 2024.

Oleg Sydorov
*/
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const DelimiterByte = byte(46)

func fastSplit(s string) []string {
	n := 1
	c := DelimiterByte

	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}

	out := make([]string, n)
	count := 0
	begin := 0
	length := len(s) - 1

	for i := 0; i <= length; i++ {
		if s[i] == c {
			out[count] = s[begin:i]
			count++
			begin = i + 1
		}
	}
	out[count] = s[begin : length+1]

	return out
}

func ipToUint32(ipStr string) (uint32, error) {
	parts := fastSplit(ipStr)
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP format")
	}

	var ipNum uint32
	for i := 0; i < 4; i++ {
		val, err := strconv.Atoi(parts[i])
		if err != nil || val < 0 || val > 255 {
			return 0, fmt.Errorf("invalid IP octet: %v", parts[i])
		}
		ipNum = (ipNum << 8) | uint32(val)
	}

	return ipNum, nil
}

func deduplicate(arr []uint32) []uint32 {
	if len(arr) == 0 {
		return arr
	}
	w := 1
	for r := 1; r < len(arr); r++ {
		if arr[r] != arr[r-1] {
			arr[w] = arr[r]
			w++
		} else {
			atomic.AddUint64(&DeletedInChunks, 1)
		}
	}
	return arr[:w]
}

func chkClose(c io.Closer) {
	if err := c.Close(); err != nil {
		fmt.Println("Error:", err)
	}
}

func worker(id int, jobs <-chan struct {
	chunk      []uint32
	chunkIndex int
}, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {
		_, err := processChunk(job.chunk, job.chunkIndex)
		if err != nil {
			log.Fatalf("Error processing chunk %d: %v, worker: %d", job.chunkIndex, err, id)
		}
	}
}

func roundToNearest1024(x int64) int64 {
	if x <= 0 {
		return 0
	}
	const step = 1024
	remainder := x % step
	if remainder == 0 {
		return x
	}
	if remainder >= step/2 {
		return x + (step - remainder)
	}
	return x - remainder
}

func suggestChunkSize(fileName string, userChunkSize int64) int64 {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()
	fmt.Printf("File size:: %v bytes (%.2f GB)\n", fileSize, float64(fileSize)/(1024*1024*1024))

	if fileSize/userChunkSize > 150 {
		fmt.Println("Warning: chosen chunk size is too small!")

		suggestedChunkSize := fileSize / 100

		rounded := roundToNearest1024(suggestedChunkSize)
		fmt.Printf("Suggested chunk size: %.2f MB. Use it? (y/n): ", float64(suggestedChunkSize)/(1024*1024))
		var response string
		_, err = fmt.Scanln(&response)
		if err != nil {
			log.Fatalf("Error reading response: %v", err)
		}
		if response != "y" {
			fmt.Println("Using original chunk size.")
			return userChunkSize
		}

		return rounded
	}

	return userChunkSize
}

func cleanUpChunk(files ...string) {
	for _, file := range files {
		err := os.Remove(file)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}

func makeTmpDir(p string) {
	err := os.Mkdir(p, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Printf("Directory %s created successfully!\n", p)
}

func cleanUpAll(p string) {
	err := os.RemoveAll(p)
	if err != nil {
		fmt.Println("Error cleaning up temporary directory:", err)
	}
}

func spinner(delay time.Duration, stop chan struct{}) {
	chars := []rune{'|', '/', '-', '\\'}
	for {
		select {
		case <-stop:
			return
		default:
			for _, c := range chars {
				select {
				case <-stop:
					fmt.Print("\r" + " " + "\r")
					return
				default:
					fmt.Printf("\r%c", c)
					time.Sleep(delay)
				}
			}
		}
	}
}
