/*
  - Copyright (c) 2024.

Oleg Sydorov
*/
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func multiPhaseMerge(files []string) (string, error) {
	fmt.Print("\r" + " " + "\r")
	fmt.Println("Merging & deduplicating...")

	stopSpinner := make(chan struct{})
	go spinner(200*time.Millisecond, stopSpinner)
	defer close(stopSpinner)
	if len(files) == 1 {
		return files[0], nil
	}

	fileIndex := 0

	results := make(chan string, *numWorkers*2)

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsMerge {
				if err := mergeTwoFiles(job.fileA, job.fileB, job.outFile, true); err != nil {
					log.Fatalf("Error merging files %s and %s: %v", job.fileA, job.fileB, err)
				}
				results <- job.outFile
			}
		}()
	}

	for len(files) > 1 {
		var newFiles []string

		for i := 0; i < len(files); i += 2 {
			if i+1 < len(files) {
				outFile := fmt.Sprintf(path+"merged_%d.tmp", fileIndex)
				fileIndex++
				jobsMerge <- struct {
					fileA, fileB string
					outFile      string
				}{fileA: files[i], fileB: files[i+1], outFile: outFile}
			} else {
				newFiles = append(newFiles, files[i])
			}
		}

		for range files[:len(files)/2] {
			newFiles = append(newFiles, <-results)
		}

		files = newFiles
	}

	close(jobsMerge)
	wg.Wait()
	close(results)

	return files[0], nil
}

func mergeTwoFiles(fileA, fileB, outFile string, cleanUp bool) error {
	if cleanUp {
		defer cleanUpChunk(fileA, fileB)
	}
	a, _ := strings.CutPrefix(fileA, path)
	b, _ := strings.CutPrefix(fileB, path)
	fmt.Print("\r" + " " + "\r")
	fmt.Printf("%s & %s\n", a, b)

	fa, err := os.Open(fileA)
	if err != nil {
		return err
	}
	defer chkClose(fa)

	fb, err := os.Open(fileB)
	if err != nil {
		return err
	}
	defer chkClose(fb)

	fOut, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer chkClose(fOut)

	const batchSize = 1024
	bufferA := make([]uint32, batchSize)
	bufferB := make([]uint32, batchSize)

	bw := bufio.NewWriter(fOut)
	defer func(bw *bufio.Writer) {
		err := bw.Flush()
		if err != nil {
			log.Fatalf("Error flushing buffer: %v", err)
		}
	}(bw)

	var indexA, sizeA, indexB, sizeB int
	var lastWritten uint32
	var hasLast bool

	readNextBatch := func(file *os.File, buffer []uint32) (int, error) {
		tempBuffer := make([]byte, len(buffer)*4)
		n, err := file.Read(tempBuffer)
		if err != nil && err != io.EOF {
			return 0, err
		}
		count := n / 4
		for i := 0; i < count; i++ {
			buffer[i] = binary.LittleEndian.Uint32(tempBuffer[i*4 : (i+1)*4])
		}
		return count, nil
	}

	sizeA, err = readNextBatch(fa, bufferA)
	if err != nil {
		return err
	}
	sizeB, err = readNextBatch(fb, bufferB)
	if err != nil {
		return err
	}

	for indexA < sizeA || indexB < sizeB {
		var valA, valB uint32
		hasA := indexA < sizeA
		hasB := indexB < sizeB

		if hasA {
			valA = bufferA[indexA]
		}
		if hasB {
			valB = bufferB[indexB]
		}

		if hasA && (!hasB || valA < valB) {
			if !hasLast || valA != lastWritten {
				err = binary.Write(bw, binary.LittleEndian, valA)
				if err != nil {
					return err
				}
				lastWritten = valA
				hasLast = true
			}
			indexA++
			if indexA == sizeA {
				sizeA, err = readNextBatch(fa, bufferA)
				if err != nil && err != io.EOF {
					return err
				}
				indexA = 0
			}
		} else if hasB && (!hasA || valB < valA) {
			if !hasLast || valB != lastWritten {
				err = binary.Write(bw, binary.LittleEndian, valB)
				if err != nil {
					return err
				}
				lastWritten = valB
				hasLast = true
			}
			indexB++
			if indexB == sizeB {
				sizeB, err = readNextBatch(fb, bufferB)
				if err != nil && err != io.EOF {
					return err
				}
				indexB = 0
			}
		} else { // <= always true
			if !hasLast || valA != lastWritten {
				err = binary.Write(bw, binary.LittleEndian, valA)
				if err != nil {
					return err
				}
				lastWritten = valA
				hasLast = true
				atomic.AddUint64(&DeletedInMerge, 1)
			} else {
				atomic.AddUint64(&DeletedInMerge, 1)
			}
			indexA++
			indexB++
			if indexA == sizeA {
				sizeA, err = readNextBatch(fa, bufferA)
				if err != nil && err != io.EOF {
					return err
				}
				indexA = 0
			}
			if indexB == sizeB {
				sizeB, err = readNextBatch(fb, bufferB)
				if err != nil && err != io.EOF {
					return err
				}
				indexB = 0
			}
		}
	}
	return nil
}
