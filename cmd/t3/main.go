package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type codeRow struct {
	code          string
	description   string
	validStartDos string
	validEndDos   string
}

func main() {
	fileName := os.Args[1]
	outputFileName := os.Args[2]
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("unable to open file: %s (%v)", fileName, err)
	}
	defer file.Close()

	var codeRows []codeRow
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, " ")
		icd10Code := parts[0]
		icd10Description := strings.TrimSpace(strings.Join(parts[1:], " "))
		codeRows = append(codeRows, codeRow{
			code:          icd10Code,
			description:   icd10Description,
			validStartDos: "2024-01-01",
			validEndDos:   "2027-12-31",
		})
	}

	header := strings.Join([]string{
		"code",
		"description",
		"valid_start_dos",
		"valid_end_dos",
	}, ",")

	outputFile, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("unable to open output file %s (%v)", outputFileName, err)
	}
	defer outputFile.Close()

	fmt.Fprintf(outputFile, "%s\n", header)
	for _, val := range codeRows {
		row := strings.Join([]string{
			val.code,
			fmt.Sprintf("\"%s\"", val.description),
			val.validStartDos,
			val.validEndDos,
		}, ",")
		fmt.Fprintf(outputFile, "%s\n", row)
	}

	outputFile.Sync()
}
