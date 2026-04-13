package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strings"
)

const techxsortVersion = "0.0.2"

func printtechxsortVersion() {
	fmt.Printf("techxsort version %s\n", techxsortVersion)
}

type TFRecord struct {
	Host  string   `json:"host"`
	Count int      `json:"count"`
	Tech  []string `json:"tech"`
}

// indexedRecord keeps original position for stable output ordering
type indexedRecord struct {
	index  int
	record TFRecord
}

// extractDomain returns the normalized hostname from a URL.
// Strips www. prefix and default ports (:80 for http, :443 for https)
// so that globalsignin.com, www.globalsignin.com, www.globalsignin.com:80
// all normalize to "globalsignin.com".
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		parsed, err = url.Parse("https://" + rawURL)
		if err != nil || parsed.Host == "" {
			return rawURL
		}
	}

	host := parsed.Hostname() // strips port
	port := parsed.Port()

	// Strip www. prefix
	host = strings.TrimPrefix(host, "www.")

	// Keep non-default ports
	if port != "" && !(port == "80" && parsed.Scheme == "http") && !(port == "443" && parsed.Scheme == "https") {
		host = host + ":" + port
	}

	return host
}

func main() {
	output := flag.String("o", "", "Path to save the output file")
	versionFlag := flag.Bool("version", false, "Print the version of the tool and exit.")
	flag.Parse()

	if *versionFlag {
		printtechxsortVersion()
		return
	}

	// 1. Parse all JSON objects from stdin
	var allRecords []indexedRecord
	decoder := json.NewDecoder(bufio.NewReader(os.Stdin))
	idx := 0
	for {
		var rec TFRecord
		if err := decoder.Decode(&rec); err != nil {
			break
		}
		if rec.Host == "" || len(rec.Tech) == 0 {
			idx++
			continue
		}
		allRecords = append(allRecords, indexedRecord{index: idx, record: rec})
		idx++
	}

	if len(allRecords) == 0 {
		return
	}

	// 2. Group records by domain
	domainGroups := make(map[string][]int) // domain -> indices into allRecords
	for i, ir := range allRecords {
		domain := extractDomain(ir.record.Host)
		domainGroups[domain] = append(domainGroups[domain], i)
	}

	// 3. For each domain group, deduplicate techs
	// Track which records survive and their filtered techs
	survivingTech := make(map[int][]string) // allRecords index -> filtered tech list
	removed := make(map[int]bool)

	for _, indices := range domainGroups {
		// Sort indices by count DESC, then by original index ASC (first-seen tiebreak)
		sorted := make([]int, len(indices))
		copy(sorted, indices)
		sort.SliceStable(sorted, func(a, b int) bool {
			return allRecords[sorted[a]].record.Count > allRecords[sorted[b]].record.Count
		})

		// Process in sorted order, claiming techs
		claimed := make(map[string]bool)
		for _, i := range sorted {
			var remaining []string
			for _, tech := range allRecords[i].record.Tech {
				if !claimed[tech] {
					claimed[tech] = true
					remaining = append(remaining, tech)
				}
			}
			if len(remaining) == 0 {
				removed[i] = true
			} else {
				survivingTech[i] = remaining
			}
		}
	}

	// 4. Prepare output file if specified
	var outputFile *os.File
	if *output != "" {
		var err error
		outputFile, err = os.OpenFile(*output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
			os.Exit(1)
		}
		defer outputFile.Close()
	}

	// 5. Output surviving records in original input order
	for i, ir := range allRecords {
		if removed[i] {
			continue
		}

		rec := ir.record
		if tech, ok := survivingTech[i]; ok {
			rec.Tech = tech
			rec.Count = len(tech)
		}

		outputJSON, err := json.MarshalIndent(rec, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling JSON for host %s: %v\n", rec.Host, err)
			os.Exit(1)
		}

		fmt.Println(string(outputJSON))

		if outputFile != nil {
			_, err := outputFile.Write(append(outputJSON, '\n'))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing to output file: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
