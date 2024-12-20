/*
 * Copyright (c) 2024.
 */

package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"
)

func TestProcessChunk(t *testing.T) {
	ips := []uint32{12345, 12345, 67890, 12345, 67890, 11111}
	expected := []uint32{11111, 12345, 67890}

	outputFile, err := processChunk(ips, 0)
	if err != nil {
		t.Fatalf("processChunk failed: %v", err)
	}
	defer func(name string) {
		err = os.Remove(name)
		if err != nil {
			fmt.Printf("error removing file: %v", err)
		}
	}(outputFile)

	output, err := readBinaryFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(output) != len(expected) {
		t.Fatalf("Expected %d IPs, got %d", len(expected), len(output))
	}
	for i, ip := range output {
		if ip != expected[i] {
			t.Errorf("Expected IP %d at index %d, got %d", expected[i], i, ip)
		}
	}
}

func TestMergeTwoFiles(t *testing.T) {
	cleanup := func(f *os.File) {
		err := f.Close()
		if err != nil {
			t.Logf("Error closing file: %v", err)
		}
		err = os.Remove(f.Name())
		if err != nil {
			t.Logf("Error removing file: %v", err)
		}
	}

	fileA, err := os.CreateTemp("", "fileA_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary fileA: %v", err)
	}
	defer cleanup(fileA)

	fileB, err := os.CreateTemp("", "fileB_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary fileB: %v", err)
	}
	defer cleanup(fileB)

	outFile, err := os.CreateTemp("", "merged_*.tmp")
	if err != nil {
		t.Fatalf("Failed to create temporary output file: %v", err)
	}
	defer cleanup(outFile)

	err = writeBinaryFile(fileA.Name(), []uint32{11111, 12345, 67890})
	if err != nil {
		t.Fatalf("Failed to write to fileA: %v", err)
	}

	err = writeBinaryFile(fileB.Name(), []uint32{12345, 23456, 67890})
	if err != nil {
		t.Fatalf("Failed to write to fileB: %v", err)
	}

	err = mergeTwoFiles(fileA.Name(), fileB.Name(), outFile.Name(), false)
	if err != nil {
		t.Fatalf("mergeTwoFiles failed: %v", err)
	}

	expected := []uint32{11111, 12345, 23456, 67890}
	output, err := readBinaryFile(outFile.Name())
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(output) != len(expected) {
		t.Fatalf("Expected %d elements, got %d", len(expected), len(output))
	}
	for i, ip := range output {
		if ip != expected[i] {
			t.Errorf("Expected %d at index %d, got %d", expected[i], i, ip)
		}
	}
}

func TestDeduplicate(t *testing.T) {
	ips := []uint32{11111, 12345, 12345, 12345, 67890, 67890}
	expected := []uint32{11111, 12345, 67890}

	result := deduplicate(ips)

	fmt.Println("Result:", result)

	if len(result) != len(expected) {
		t.Fatalf("Expected %d unique IPs, got %d", len(expected), len(result))
	}
	for i, ip := range result {
		if ip != expected[i] {
			t.Errorf("Expected IP %d at index %d, got %d", expected[i], i, ip)
		}
	}
}

func readBinaryFile(fileName string) ([]uint32, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer chkClose(f)

	var result []uint32
	for {
		var ip uint32
		err := binary.Read(f, binary.LittleEndian, &ip)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}
		result = append(result, ip)
	}

	return result, nil
}

func writeBinaryFile(fileName string, data []uint32) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer chkClose(f)

	for _, ip := range data {
		err := binary.Write(f, binary.LittleEndian, ip)
		if err != nil {
			return err
		}
	}
	return nil
}
