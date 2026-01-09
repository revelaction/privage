package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/revelaction/privage/header"
	"github.com/revelaction/privage/setup"
)

// listCommand list encripted files
func listCommand(s *setup.Setup, filter string, ui UI) error {
	headers := []*header.Header{}
	failures := []*header.Header{}

	if s.Id.Id == nil {
		return fmt.Errorf("%w: %v", ErrNoIdentity, s.Id.Err)
	}

	ch, err := headerGenerator(s.Repository, s.Id)
	if err != nil {
		return err
	}

	for h := range ch {
		if h.Err != nil {
			failures = append(failures, h)
		} else {
			headers = append(headers, h)
		}
	}

	var toList, toListForCat, toListForLabel []*header.Header
	if filter == "" {
		toList = headers
		_, _ = fmt.Fprintf(ui.Out, "Found %d total encrypted tracked files.\n", len(toList))
		sorted := sortList(toList)
		for _, h := range sorted {
			_, _ = fmt.Fprintf(ui.Out, "%8s%s\n", "", h)
		}
	} else {

		toListForCat = headersForFilterCat(filter, headers)
		toListForLabel = headersForFilterLabel(filter, headers)

		if len(toListForCat) == 0 && len(toListForLabel) == 0 {
			_, _ = fmt.Fprintf(ui.Out, "Found no encrypted tracked files matching '%s'\n", filter)
		}

		if len(toListForCat) > 0 {
			_, _ = fmt.Fprintf(ui.Out, "Found %d files with category matching '%s':\n", len(toListForCat), filter)
			_, _ = fmt.Fprintln(ui.Out)
			sorted := sortList(toListForCat)
			for _, h := range sorted {
				_, _ = fmt.Fprintf(ui.Out, "%8s%s\n", "", h)
			}

			_, _ = fmt.Fprintln(ui.Out)
		}

		if len(toListForLabel) > 0 {
			_, _ = fmt.Fprintf(ui.Out, "Found %d files with name matching '%s':\n", len(toListForLabel), filter)
			_, _ = fmt.Fprintln(ui.Out)
			sorted := sortList(toListForLabel)
			for _, h := range sorted {
				_, _ = fmt.Fprintf(ui.Out, "%8s%s\n", "", h)
			}

			_, _ = fmt.Fprintln(ui.Out)
		}
	}

	if len(failures) > 0 {
		_, _ = fmt.Fprintf(ui.Out, "\nFound %d files with errors:\n", len(failures))
		for _, f := range failures {
			name := filepath.Base(f.Path)
			_, _ = fmt.Fprintf(ui.Out, "💥 %s Error: %v\n", name, f.Err)
		}
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
