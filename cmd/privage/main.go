// Copyright (c) 2022 The privage developers

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/revelaction/privage/setup"
)

var (
	BuildCommit    string
	BuildTag       string
	YubikeySupport string
)

// UI contains the output streams for the application.
// Used for injecting buffers during testing.
type UI struct {
	Out io.Writer
	Err io.Writer
}

func main() {
	ui := UI{Out: os.Stdout, Err: os.Stderr}

	cmd, args, global, err := parseMainArgs(os.Args[1:], ui)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		// Error message already printed by parseMainArgs to ui.Err
		os.Exit(1)
	}

	if err := runCommand(cmd, args, global, ui); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		FprintErr(ui.Err, err)
		os.Exit(1)
	}
}

func FprintErr(w io.Writer, err error) {
	_, _ = fmt.Fprintf(w, "privage: %v\n", err)
}

func parseMainArgs(args []string, ui UI) (string, []string, setup.Options, error) {
	fs := flag.NewFlagSet("privage", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	setupUsage(fs)

	var opts setup.Options

	fs.StringVar(&opts.ConfigFile, "conf", "", "Use file as privage configuration file")
	fs.StringVar(&opts.ConfigFile, "c", "", "alias for -conf")
	fs.StringVar(&opts.KeyFile, "key", "", "Use file path for private key")
	fs.StringVar(&opts.KeyFile, "k", "", "alias for -key")
	fs.StringVar(&opts.PivSlot, "piv-slot", "", "The PIV slot for decryption of the age key")
	fs.StringVar(&opts.PivSlot, "p", "", "alias for -piv-slot")
	fs.StringVar(&opts.RepoPath, "repository", "", "Use file path as path for the encrypted files")
	fs.StringVar(&opts.RepoPath, "r", "", "alias for -repository")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", nil, opts, err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
		fs.Usage()
		return "", nil, opts, err
	}

	if fs.NArg() == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", nil, opts, errors.New("no command provided")
	}

	cmd := fs.Arg(0)
	cmdArgs := fs.Args()[1:]
	return cmd, cmdArgs, opts, nil
}

func runCommand(cmd string, args []string, opts setup.Options, ui UI) error {

	switch cmd {
	// 1. Utility commands (No setup needed)
	case "version":
		return versionCommand(ui)
	case "bash":
		if err := parseBashArgs(args, ui); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		return bashCommand(ui)
	case "complete":
		return completeCommand(opts, args, ui) // needs raw opts for sub-dispatch
	case "help":
		if len(args) > 0 {
			return runCommand(args[0], []string{"--help"}, opts, ui)
		}
		// For general help, we show the main usage to ui.Out
		fs := flag.NewFlagSet("privage", flag.ContinueOnError)
		fs.SetOutput(ui.Out)
		setupUsage(fs)
		fs.Usage()
		return nil

	// 2. Bootstrap commands (Needs raw Options, not Setup)
	case "init":
		slot, err := parseInitArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		return initCommand(slot, ui)

	// 3. Operational commands (Require full Setup)
	case "cat":
		label, err := parseCatArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}

		return catCommand(s, label, ui)

	case "add":
		cat, label, err := parseAddArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}

		return addCommand(s, cat, label, ui)

	case "show":
		label, fieldName, err := parseShowArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}

		return showCommand(s, label, fieldName, ui)

	case "delete":
		label, err := parseDeleteArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}

		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}

		return deleteCommand(s, label, ui)

	case "key":
		if err := parseKeyArgs(args, ui); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return keyCommand(s, ui)

	case "status":
		if err := parseStatusArgs(args, ui); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return statusCommand(s, ui)

	case "list":
		filter, err := parseListArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return listCommand(s, filter, ui)

	case "clipboard":
		label, err := parseClipboardArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return clipboardCommand(s, label, ui)

	case "decrypt":
		label, err := parseDecryptArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return decryptCommand(s, label, ui)

	case "reencrypt":
		force, clean, err := parseReencryptArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return reencryptCommand(s, force, clean, ui)

	case "rotate":
		clean, slot, err := parseRotateArgs(args, ui)
		if err != nil {
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}
			return err
		}
		s, setupErr := setupEnv(opts)
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}
		return rotateCommand(s, clean, slot, ui)
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

func setupUsage(fs *flag.FlagSet) {
	fs.Usage = func() {
		output := fs.Output()
		_, _ = fmt.Fprintf(output, "Usage: %s [global options] command [command options] [arguments...]\n", os.Args[0])
		_, _ = fmt.Fprintf(output, "\nCommands:\n")
		_, _ = fmt.Fprintf(output, "  init       Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		_, _ = fmt.Fprintf(output, "  key        Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
		_, _ = fmt.Fprintf(output, "  status     Provide information about the current configuration.\n")
		_, _ = fmt.Fprintf(output, "  add        Add a new encrypted file.\n")
		_, _ = fmt.Fprintf(output, "  delete     Delete an encrypted file.\n")
		_, _ = fmt.Fprintf(output, "  list       list metadata of all/some encrypted files.\n")
		_, _ = fmt.Fprintf(output, "  show       Show the contents the an encripted file.\n")
		_, _ = fmt.Fprintf(output, "  cat        Print the full contents of an encrypted file to stdout.\n")
		_, _ = fmt.Fprintf(output, "  clipboard  Copy the credential password to the clipboard\n")
		_, _ = fmt.Fprintf(output, "  decrypt    Decrypt a file and write its content in a file named after the label\n")
		_, _ = fmt.Fprintf(output, "  reencrypt  Reencrypt all decrypted files that are already encrypted. (default is dry-run)\n")
		_, _ = fmt.Fprintf(output, "  rotate     Create a new age key and reencrypt every file with the new key\n")
		_, _ = fmt.Fprintf(output, "  bash       Dump bash complete script.\n")
		_, _ = fmt.Fprintf(output, "  version    Show version information\n")
		_, _ = fmt.Fprintf(output, "  help       Show help for a command.\n")
		_, _ = fmt.Fprintf(output, "\nGlobal Options:\n")
		_, _ = fmt.Fprintf(output, "  -h, --help             Show help for privage\n")
		_, _ = fmt.Fprintf(output, "  -c, -conf string       Use file as privage configuration file\n")
		_, _ = fmt.Fprintf(output, "  -k, -key string        Use file path for private key\n")
		_, _ = fmt.Fprintf(output, "  -p, -piv-slot string   The PIV slot for decryption of the age key\n")
		_, _ = fmt.Fprintf(output, "  -r, -repository string Use file path as path for the encrypted files\n")
		_, _ = fmt.Fprintf(output, "\nVersion: %s, commit %s, yubikey %s\n", BuildTag, BuildCommit, YubikeySupport)
	}
}
