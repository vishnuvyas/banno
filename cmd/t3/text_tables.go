package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type TextTableFile struct {
	filename   string
	seperator  string
	coalasce   bool
	maxColumns int
	file       *os.File
}

type TTCursor struct {
	RowNum  int
	scanner *bufio.Scanner
	ttFile  *TextTableFile
}

type Row struct {
	Cells  []string
	RowNum int
}

func (t *TextTableFile) Close() error {
	return t.file.Close()
}

func Open(filename string, separator string, maxColumns int, coalasce bool) (*TextTableFile, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %w", err)
	}

	return &TextTableFile{
		filename:   filename,
		seperator:  separator,
		maxColumns: maxColumns,
		coalasce:   coalasce,
		file:       file,
	}, nil
}

func (t *TextTableFile) Cursor() (*TTCursor, error) {
	scanner := bufio.NewScanner(t.file)
	return &TTCursor{
		RowNum:  -1,
		scanner: scanner,
		ttFile:  t,
	}, nil
}

func (c *TTCursor) parseRow() Row {
	line := strings.TrimSpace(c.scanner.Text())
	parts := strings.Split(line, c.ttFile.seperator)
	var cells []string
	if c.ttFile.maxColumns > 0 {
		validCells := parts[:c.ttFile.maxColumns-1]
		remainingCells := parts[c.ttFile.maxColumns-1:]
		finalCell := strings.TrimSpace(strings.Join(remainingCells, c.ttFile.seperator))
		cells = validCells
		cells = append(cells, finalCell)
	}

	return Row{
		Cells:  cells,
		RowNum: c.RowNum,
	}
}

func (c *TTCursor) HasNext() bool {
	return c.scanner.Scan()
}

func (c *TTCursor) Next() Row {
	c.RowNum += 1
	return c.parseRow()
}
