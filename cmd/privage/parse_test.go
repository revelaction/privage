package main

import (
	"bytes"
	"errors"
	"flag"
	"strings"
	"testing"
)

func TestParseCatArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, err := parseCatArgs([]string{"mylabel"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "mylabel" {
			t.Errorf("got label %q, want %q", label, "mylabel")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseCatArgs([]string{"-h"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseCatArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})

	t.Run("MissingLabel", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseCatArgs([]string{}, ui)
		if err == nil {
			t.Fatal("expected error for missing label")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseInitArgs(t *testing.T) {
	t.Run("SuccessEmpty", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		slot, err := parseInitArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slot != "" {
			t.Errorf("got slot %q, want empty", slot)
		}
	})

	t.Run("SuccessSlot", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		slot, err := parseInitArgs([]string{"-p", "9c"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if slot != "9c" {
			t.Errorf("got slot %q, want %q", slot, "9c")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseInitArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseInitArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseAddArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		cat, lab, err := parseAddArgs([]string{"cred", "mylabel"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cat != "cred" || lab != "mylabel" {
			t.Errorf("got %q/%q, want %q/%q", cat, lab, "cred", "mylabel")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseAddArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("MissingArgs", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseAddArgs([]string{"cred"}, ui)
		if err == nil {
			t.Fatal("expected error for missing argument")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})

	t.Run("CategoryTooLong", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseAddArgs([]string{strings.Repeat("a", 33), "lab"}, ui)
		if err == nil {
			t.Fatal("expected error for long category")
		}
	})

	t.Run("LabelTooLong", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseAddArgs([]string{"cat", strings.Repeat("a", 129)}, ui)
		if err == nil {
			t.Fatal("expected error for long label")
		}
	})
}

func TestParseShowArgs(t *testing.T) {
	t.Run("LabelOnly", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, field, err := parseShowArgs([]string{"my"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "my" || field != "" {
			t.Errorf("got %q/%q, want %q/empty", label, field, "my")
		}
	})

	t.Run("LabelAndField", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, field, err := parseShowArgs([]string{"my", "pass"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "my" || field != "pass" {
			t.Errorf("got %q/%q, want %q/%q", label, field, "my", "pass")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseShowArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("MissingLabel", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseShowArgs([]string{}, ui)
		if err == nil {
			t.Fatal("expected error for missing label")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseReencryptArgs(t *testing.T) {
	t.Run("SuccessDryRun", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		force, clean, err := parseReencryptArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if force || clean {
			t.Errorf("got force=%v, clean=%v, want both false", force, clean)
		}
	})

	t.Run("SuccessForceAndClean", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		force, clean, err := parseReencryptArgs([]string{"-f", "-c"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !force || !clean {
			t.Errorf("got force=%v, clean=%v, want both true", force, clean)
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseReencryptArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseReencryptArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseRotateArgs(t *testing.T) {
	t.Run("SuccessDefault", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		clean, slot, err := parseRotateArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if clean || slot != "" {
			t.Errorf("got clean=%v, slot=%q, want false/empty", clean, slot)
		}
	})

	t.Run("SuccessCleanAndSlot", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		clean, slot, err := parseRotateArgs([]string{"--clean", "--piv-slot", "9e"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !clean || slot != "9e" {
			t.Errorf("got clean=%v, slot=%q, want true/9e", clean, slot)
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseRotateArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, _, err := parseRotateArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseListArgs(t *testing.T) {
	t.Run("SuccessNoFilter", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		filter, err := parseListArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filter != "" {
			t.Errorf("got filter %q, want empty", filter)
		}
	})

	t.Run("SuccessFilter", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		filter, err := parseListArgs([]string{"myfilter"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if filter != "myfilter" {
			t.Errorf("got filter %q, want %q", filter, "myfilter")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseListArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseListArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseDeleteArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, err := parseDeleteArgs([]string{"mylabel"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "mylabel" {
			t.Errorf("got label %q, want %q", label, "mylabel")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseDeleteArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("MissingLabel", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseDeleteArgs([]string{}, ui)
		if err == nil {
			t.Fatal("expected error for missing label")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseKeyArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseKeyArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseKeyArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseKeyArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseStatusArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseStatusArgs([]string{}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseStatusArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("UnknownFlag", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		err := parseStatusArgs([]string{"--foo"}, ui)
		if err == nil {
			t.Fatal("expected error for unknown flag")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseClipboardArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, err := parseClipboardArgs([]string{"mylabel"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "mylabel" {
			t.Errorf("got label %q, want %q", label, "mylabel")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseClipboardArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("MissingLabel", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseClipboardArgs([]string{}, ui)
		if err == nil {
			t.Fatal("expected error for missing label")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}

func TestParseDecryptArgs(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		label, err := parseDecryptArgs([]string{"mylabel"}, ui)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if label != "mylabel" {
			t.Errorf("got label %q, want %q", label, "mylabel")
		}
	})

	t.Run("Help", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseDecryptArgs([]string{"--help"}, ui)
		if !errors.Is(err, flag.ErrHelp) {
			t.Fatalf("expected flag.ErrHelp, got %v", err)
		}
		if outBuf.Len() == 0 {
			t.Error("expected usage output in Out buffer")
		}
	})

	t.Run("MissingLabel", func(t *testing.T) {
		var outBuf, errBuf bytes.Buffer
		ui := UI{Out: &outBuf, Err: &errBuf}
		_, err := parseDecryptArgs([]string{}, ui)
		if err == nil {
			t.Fatal("expected error for missing label")
		}
		if errBuf.Len() == 0 {
			t.Error("expected error message in Err buffer")
		}
	})
}