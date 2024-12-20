# IP-Addr-Counter
##### Calculates the number of unique addresses in a given fileCalculates the number of unique addresses in a given file

**Usage:**
    -size int
Enter chunk size (default 1073741824)****
    -w int
Enter worker pool size (default 16)

For huge test file, recommended options are:

**.\ip_calc -size 5368709120 -w 4**

Keep balance between chunk size and amount of workers.
This allows you to adjust needed CPU/Memory consumption balance.

#### For testing purposes, please build the generatorFor testing purposes, please build the generator

**Usage of ./generator:**
-a int
Enter amount of IPs (default 100)

#### For more details, explore the Makefile:

make [target]

**Targets:**
- all          Build the project (default target)
- build        Build the binary for the current system
- build-linux  Build the binary for Linux/amd64
- run          Build and run the project
- generator    Build and run the test file generator
- lint         Run the linter (requires golangci-lint)
- test         Run all tests
- clean        Remove build artifacts
- help         Show this help message


With the best regards, 
**Oleg Sydorov**

Enjoy!