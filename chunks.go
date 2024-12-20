/*
  - Copyright (c) 2024.

Oleg Sydorov
*/
package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"sort"
	"time"
)

func processChunk(ips []uint32, chunkIndex int) (string, error) {
	fmt.Print("\r" + " " + "\r")
	fmt.Printf("Processing chunk %d ...sorting, transforming to binary\n", chunkIndex)
	stopSpinner := make(chan struct{})
	go spinner(200*time.Millisecond, stopSpinner)
	defer close(stopSpinner)

	sort.Slice(ips, func(i, j int) bool { return ips[i] < ips[j] })

	ips = deduplicate(ips)

	outFileName := fmt.Sprintf("chunk_%d.tmp", chunkIndex)

	f, err := os.Create(path + outFileName)
	if err != nil {
		return "", err
	}
	defer chkClose(f)

	bw := bufio.NewWriter(f)

	for _, ip := range ips {
		err := binary.Write(bw, binary.LittleEndian, ip)
		if err != nil {
			return "", fmt.Errorf("failed to write binary data: %w", err)
		}
	}

	err = bw.Flush()
	if err != nil {
		return "", fmt.Errorf("failed to flush buffer: %w", err)
	}

	chunkFiles = append(chunkFiles, path+outFileName)

	return outFileName, nil
}
