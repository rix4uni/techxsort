package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

type Record struct {
	Host  string   `json:"host"`
	Count int      `json:"count"`
	Tech  []string `json:"tech"`
}

func main() {
	// Define command-line flags
	file := flag.String("file", "", "Path to the techx JSON file")
	output := flag.String("o", "", "Path to save the output file")
	flag.Parse()

	// Check if file flag is provided
	if *file == "" {
		fmt.Println("Please specify the JSON file using -file flag.")
		os.Exit(1)
	}

	// Open the JSON file
	jsonFile, err := os.Open(*file)
	if err != nil {
		fmt.Printf("Error opening JSON file: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()

	// Use buffered reader for reading domains from stdin
	scanner := bufio.NewScanner(os.Stdin)
	var inputDomains []string
	for scanner.Scan() {
		inputDomains = append(inputDomains, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		os.Exit(1)
	}

	// Load records concurrently
	var records []Record
	var recordMutex sync.Mutex
	var wg sync.WaitGroup

	decoder := json.NewDecoder(bufio.NewReader(jsonFile))
	for {
		var record Record
		if err := decoder.Decode(&record); err != nil {
			break
		}
		wg.Add(1)
		go func(record Record) {
			defer wg.Done()
			recordMutex.Lock()
			records = append(records, record)
			recordMutex.Unlock()
		}(record)
	}
	wg.Wait()

	// Process each domain and match against JSON records
	results := make(map[string]*Record)
	var resultsMutex sync.Mutex
	wg = sync.WaitGroup{}

	// Process each domain concurrently
	for _, domain := range inputDomains {
		wg.Add(1)
		go func(domain string) {
			defer wg.Done()
			uniqueTech := make(map[string]struct{})
			totalCount := 0

			// Find matching records for each domain
			for _, record := range records {
				if strings.HasPrefix(record.Host, domain) {
					
					// Collect unique technologies
					for _, tech := range record.Tech {
						uniqueTech[tech] = struct{}{}
					}
					totalCount += record.Count
				}
			}

			// Only add to results if there are matching records
			if len(uniqueTech) > 0 {
				techList := make([]string, 0, len(uniqueTech))
				for tech := range uniqueTech {
					techList = append(techList, tech)
				}

				// Sort technologies for consistent output
				sort.Strings(techList)

				resultsMutex.Lock()
				results[domain] = &Record{
					Host:  domain,
					Count: len(techList), // Count of unique technologies
					Tech:  techList,
				}
				resultsMutex.Unlock()
			}
		}(domain)
	}
	wg.Wait()

	// Prepare output file if specified
    var outputFile *os.File
    if *output != "" {
        var err error
        outputFile, err = os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
        if err != nil {
            fmt.Printf("Error opening output file: %v\n", err)
            os.Exit(1)
        }
        defer outputFile.Close()
    }

	// Print or save results for each domain
	for domain, result := range results {
		outputJSON, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling output JSON for domain %s: %v\n", domain, err)
			os.Exit(1)
		}

		// Print output to terminal
		fmt.Println(string(outputJSON))

		// Save output to a file
		if outputFile != nil {
			_, err := outputFile.Write(append(outputJSON, '\n'))
			if err != nil {
				fmt.Printf("Error writing to output file for domain %s: %v\n", domain, err)
				os.Exit(1)
			}
		}
	}
}
