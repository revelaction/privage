package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/revelaction/privage/header"
	//"github.com/revelaction/privage/setup"
)

// listAction list encripted files
func listAction(ctx *cli.Context) error {

	s, err := setupEnv(ctx)
	if err != nil {
		return fmt.Errorf("Unable to setup environment configuration: %s", err)
	}

	filter := "" // no filter
	if ctx.Args().Len() != 0 {
		filter = ctx.Args().First()
	}

	headers := []*header.Header{}

	if s.Id.Id == nil {
		return fmt.Errorf("Found no privage key file: %w", s.Id.Err)
	}

	for h := range headerGenerator(s.Repository, s.Id) {
		headers = append(headers, h)
	}

	var toList, toListForCat, toListForLabel []*header.Header
	if filter == "" {
		toList = headers
		fmt.Printf("Found %d total encrypted tracked files.\n", len(toList))
		sorted := sortList(toList)
		for _, h := range sorted {
			fmt.Printf("%8s%s\n", "", h)
		}

		return nil
	}

	toListForCat = headersForFilterCat(filter, headers)
	toListForLabel = headersForFilterLabel(filter, headers)

	if len(toListForCat) == 0 && len(toListForLabel) == 0 {
		fmt.Printf("Found no encrypted tracked files matching '%s'\n", filter)
		return nil
	}

	if len(toListForCat) > 0 {
		fmt.Printf("Found %d files with category matching '%s':\n", len(toListForCat), filter)
		fmt.Println()
		sorted := sortList(toListForCat)
		for _, h := range sorted {
			fmt.Printf("%8s%s\n", "", h)
		}

		fmt.Println()
	}

	if len(toListForLabel) > 0 {
		fmt.Printf("Found %d files with name matching '%s':\n", len(toListForLabel), filter)
		fmt.Println()
		sorted := sortList(toListForLabel)
		for _, h := range sorted {
			fmt.Printf("%8s%s\n", "", h)
		}

		fmt.Println()
	}

	return nil
}

func sortList(s []*header.Header) []*header.Header {
	sort.Slice(s, func(i, j int) bool {
		// Counter sorting
		if s[i].Category > s[j].Category {
			return false
		}
		if s[i].Category < s[j].Category {
			return true
		}
		if s[i].Label > s[j].Label {
			return false
		}
		if s[i].Label < s[j].Label {
			return true
		}

		return false
	})

	return s
}

func headersForFilterLabel(substring string, headers []*header.Header) []*header.Header {
	toList := []*header.Header{}
	for _, h := range headers {
		if strings.Contains(h.Label, substring) {
			toList = append(toList, h)
		}
	}

	return toList
}

func headersForFilterCat(substring string, headers []*header.Header) []*header.Header {
	toList := []*header.Header{}
	for _, h := range headers {
		if strings.Contains(h.Category, substring) {
			toList = append(toList, h)
		}
	}

	return toList
}
