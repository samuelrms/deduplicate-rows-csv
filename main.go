package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	inputName := os.Getenv("INPUT_NAME")
	if inputName == "" {
		inputName = "dados.csv"
		fmt.Println("INPUT_NAME not defined; using 'dados.csv'")
	}
	outputName := os.Getenv("OUTPUT_NAME")
	if outputName == "" {
		outputName = "dedup.csv"
		fmt.Println("OUTPUT_NAME not defined; using 'dedup.csv'")
	}
	keys := os.Getenv("KEYS")
	if keys == "" {
		keys = "empresa,moeda"
		fmt.Println("KEYS not defined; using 'empresa,moeda'")
	}

	inputPath := filepath.Join("docs", inputName)
	outputDir := "data"
	outputPath := filepath.Join(outputDir, outputName)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	inFile, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", inputPath, err)
		os.Exit(1)
	}
	defer inFile.Close()

	reader := csv.NewReader(inFile)
	reader.FieldsPerRecord = -1

	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	defer outFile.Close()
	writer := csv.NewWriter(outFile)

	header, err := reader.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
		os.Exit(1)
	}

	keyNames := strings.Split(keys, ",")
	for i := range keyNames {
		keyNames[i] = strings.TrimSpace(keyNames[i])
	}

	keyIdxs := make([]int, len(keyNames))
	for i, kn := range keyNames {
		found := false
		for j, col := range header {
			if strings.EqualFold(col, kn) {
				keyIdxs[i] = j
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "Key column '%s' not found in header\n", kn)
			os.Exit(1)
		}
	}

	if err := writer.Write(header); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing header: %v\n", err)
		os.Exit(1)
	}

	seen := make(map[string]bool)

	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading record: %v\n", err)
			os.Exit(1)
		}

		parts := make([]string, len(keyIdxs))
		for i, idx := range keyIdxs {
			if idx < len(rec) {
				parts[i] = rec[idx]
			} else {
				parts[i] = ""
			}
		}
		key := strings.Join(parts, "|")

		if !seen[key] {
			if err := writer.Write(rec); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing record: %v\n", err)
				os.Exit(1)
			}
			seen[key] = true
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Error finalizing write: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Deduplicated based on [%s] → %s\n", keys, outputPath)
}
