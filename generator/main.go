package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {

	count := flag.Int("a", 100, "Enter amount of IPs")

	flag.Parse()

	filename := "ip_addresses.txt"

	if err := generateIPAddresses(filename, *count); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Generated %d IP addresses in %s.\n", *count, filename)
	}
}

func generateIPAddresses(filename string, count int) error {
	rand.Seed(time.Now().UnixNano())

	randomIP := func() string {
		return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(256), rand.Intn(256), rand.Intn(256), rand.Intn(256))
	}

	uniqueIPs := make([]string, count/2)
	for i := 0; i < count/2; i++ {
		uniqueIPs[i] = randomIP()
	}

	finalIPs := append(uniqueIPs, uniqueIPs...)
	rand.Shuffle(len(finalIPs), func(i, j int) { finalIPs[i], finalIPs[j] = finalIPs[j], finalIPs[i] })

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, ip := range finalIPs {
		if _, err := file.WriteString(ip + "\n"); err != nil {
			return err
		}
	}
	return nil
}
