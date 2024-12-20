/*
  - Copyright (c) 2024.

Oleg Sydorov
*/
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	size            = flag.Int("size", 1<<30, "Enter chunk size")
	numWorkers      = flag.Int("w", runtime.NumCPU(), "Enter worker pool size")
	inputFileName   = flag.String("f", "ip_addresses", "Input file name")
	path            = "./tmp/"
	chunkFiles      []string
	totalLines      uint64 = 0
	DeletedInChunks uint64 = 0
	DeletedInMerge  uint64 = 0
	jobsChunk       chan struct {
		chunk      []uint32
		chunkIndex int
	}
	jobsMerge chan struct {
		fileA, fileB string
		outFile      string
	}
	wg sync.WaitGroup
)

func Init() {
	flag.Parse()

	jobsChunk = make(chan struct {
		chunk      []uint32
		chunkIndex int
	}, *numWorkers*2)
	jobsMerge = make(chan struct {
		fileA, fileB string
		outFile      string
	}, *numWorkers*2)
}

func main() {
	defer cleanUpAll(path)

	Init()

	fmt.Println("CPUs detected:", runtime.NumCPU())
	fmt.Println("Workers in pool:", *numWorkers)

	t := time.Now()

	file, err := os.Open(*inputFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer chkClose(file)

	chunkSizeBytes := suggestChunkSize(*inputFileName, int64(*size))

	makeTmpDir(path)

	scanner := bufio.NewScanner(file)

	var chunk []uint32
	var currentBytes int64
	chunkIndex := 0

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobsChunk, &wg)
	}

	fmt.Println("Reading, dividing by chunks, sorting, deduplicating...")

	for scanner.Scan() {
		line := scanner.Text()
		totalLines++
		lineSize := int64(len(line) + 1)

		ipNum, err := ipToUint32(line)
		if err != nil {
			log.Fatalf("Incorrect IP: %s\n", line)
		}

		if currentBytes+lineSize > chunkSizeBytes && currentBytes > 0 {
			cpy := make([]uint32, len(chunk))
			copy(cpy, chunk)
			jobsChunk <- struct {
				chunk      []uint32
				chunkIndex int
			}{chunk: cpy, chunkIndex: chunkIndex}
			chunkIndex++

			chunk = chunk[:0]
			currentBytes = 0
		}

		chunk = append(chunk, ipNum)
		currentBytes += lineSize
	}

	if len(chunk) > 0 {
		cpy := make([]uint32, len(chunk))
		copy(cpy, chunk)
		jobsChunk <- struct {
			chunk      []uint32
			chunkIndex int
		}{chunk: cpy, chunkIndex: chunkIndex}
	}

	close(jobsChunk)

	wg.Wait()

	if err := scanner.Err(); err != nil {
		fmt.Println("Scanner error:", err)
	}

	_, err = multiPhaseMerge(chunkFiles)
	if err != nil {
		fmt.Println("Error:", err)
	}

	finalCount := totalLines - (DeletedInChunks + DeletedInMerge)

	fmt.Print("\r" + " " + "\r")
	fmt.Println("======================================")
	fmt.Printf("Total lines: %v\n", totalLines)
	fmt.Printf("Unique IP count: %d\n", finalCount)
	fmt.Printf("Total time taken: %s\n", time.Since(t))
	fmt.Println("=======================================")
}
