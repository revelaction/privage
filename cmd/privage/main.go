// Copyright (c) 2022 The privage developers

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	filesystem "github.com/revelaction/privage/fs"
	"github.com/revelaction/privage/setup"
)

var global setup.Options

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

	flag.Usage = func() {
		output := flag.CommandLine.Output()
		fmt.Fprintf(output, "Usage: %s [global options] command [command options] [arguments...]\n", os.Args[0])
		fmt.Fprintf(output, "\nCommands:\n")
		fmt.Fprintf(output, "  init       Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		fmt.Fprintf(output, "  key        Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
		fmt.Fprintf(output, "  status     Provide information about the current configuration.\n")
		fmt.Fprintf(output, "  add        Add a new encrypted file.\n")
		fmt.Fprintf(output, "  delete     Delete an encrypted file.\n")
		fmt.Fprintf(output, "  list       list metadata of all/some encrypted files.\n")
		fmt.Fprintf(output, "  show       Show the contents the an encripted file.\n")
		fmt.Fprintf(output, "  cat        Print the full contents of an encrypted file to stdout.\n")
		fmt.Fprintf(output, "  clipboard  Copy the credential password to the clipboard\n")
		fmt.Fprintf(output, "  decrypt    Decrypt a file and write its content in a file named after the label\n")
		fmt.Fprintf(output, "  reencrypt  Reencrypt all decrypted files that are already encrypted. (default is dry-run)\n")
		fmt.Fprintf(output, "  rotate     Create a new age key and reencrypt every file with the new key\n")
		fmt.Fprintf(output, "  bash       Dump bash complete script.\n")
		fmt.Fprintf(output, "  version    Show version information\n")
		fmt.Fprintf(output, "  help, h    Show help for a command.\n")
		fmt.Fprintf(output, "\nGlobal Options:\n")
		fmt.Fprintf(output, "  -h, --help\n")
		fmt.Fprintf(output, "    \tShow help for privage\n")
		fmt.Fprintf(output, "  -c, -conf string\n")
		fmt.Fprintf(output, "    \tUse file as privage configuration file\n")
		fmt.Fprintf(output, "  -k, -key string\n")
		fmt.Fprintf(output, "    \tUse file path for private key\n")
		fmt.Fprintf(output, "  -p, -piv-slot string\n")
		fmt.Fprintf(output, "    \tThe PIV slot for decryption of the age key\n")
		fmt.Fprintf(output, "  -r, -repository string\n")
		fmt.Fprintf(output, "    \tUse file path as path for the repository\n")
		fmt.Fprintf(output, "\nVersion: %s, commit %s, yubikey %s\n", BuildTag, BuildCommit, YubikeySupport)
	}

	// Global -h/--help should go to Stdout
	flag.CommandLine.SetOutput(os.Stdout)

	flag.StringVar(&global.ConfigFile, "conf", "", "Use file as privage configuration file")
	flag.StringVar(&global.ConfigFile, "c", "", "alias for -conf")
	flag.StringVar(&global.KeyFile, "key", "", "Use file path for private key")
	flag.StringVar(&global.KeyFile, "k", "", "alias for -key")
	flag.StringVar(&global.PivSlot, "piv-slot", "", "The PIV slot for decryption of the age key")
	flag.StringVar(&global.PivSlot, "p", "", "alias for -piv-slot")
	flag.StringVar(&global.RepoPath, "repository", "", "Use file path as path for the repository")
	flag.StringVar(&global.RepoPath, "r", "", "alias for -repository")

	flag.Parse()

	if flag.NArg() == 0 {
		flag.CommandLine.SetOutput(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	if err := runCommand(cmd, args, global); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fatal(err)
	}
}

func runCommand(cmd string, args []string, opts setup.Options) error {
	ui := UI{Out: os.Stdout, Err: os.Stderr}

	switch cmd {
	// 1. Utility commands (No setup needed)
	case "version":
		return versionCommand(ui)
	case "bash":
		return bashCommand(ui)
	case "complete":
		return completeCommand(opts, args, ui) // needs raw opts for sub-dispatch
	case "help", "h":
		if len(args) > 0 {
			return runCommand(args[0], []string{"--help"}, opts)
		}
		flag.CommandLine.SetOutput(ui.Out)
		flag.Usage()
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

		// Pre-flight checks (Driver responsibility)
		configPath, err := filesystem.FindConfigFile()
		if err != nil {
			return fmt.Errorf("error searching for config file: %w", err)
		}
		if configPath != "" {
			fmt.Fprintf(ui.Err, "📑 Config file already exists: %s... Exiting\n", configPath)
			return nil
		}

		idPath, err := filesystem.FindIdentityFile()
		if err != nil {
			return fmt.Errorf("error searching for identity file: %w", err)
		}
		if idPath != "" {
			fmt.Fprintf(ui.Err, "🔑 privage key file already exists: %s... Exiting.\n", idPath)
			return nil
		}

		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}

		return initCommand(slot, currentDir, ui)

	// 3. Operational commands (Require full Setup)
	case "key", "status", "add", "delete", "list", "show", "cat", "clipboard", "decrypt", "reencrypt", "rotate":

		// POC: The "cat" command now follows the centralized driver design.
		// All CLI concerns (parsing, help, usage, positional validation) happen here.
		if cmd == "cat" {
			label, err := parseCatArgs(args, ui)
			if err != nil {
				if errors.Is(err, flag.ErrHelp) {
					return nil
				}
				return err
			}

			// 3. Resource Acquisition (The "Fat Driver" phase).
			// Only after CLI concerns are satisfied do we attempt to build the domain environment.
			s, setupErr := setupEnv(opts)
			if setupErr != nil {
				return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
			}

			// 4. Execution.
			// The subcommand is now a "clean" worker receiving pre-validated data.
			return catCommand(s, label, ui)
		}

		if cmd == "add" {
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

			// Check label exists
			if labelExists(label, s.Id) {
				return errors.New("second argument (label) already exist")
			}

			return addCommand(s, cat, label, ui)
		}

		if cmd == "show" {
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
		}

		// Legacy coordination for other commands (to be refactored next)
		s, setupErr := setupEnv(opts)

		var cmdErr error
		switch cmd {
		case "key":
			cmdErr = keyCommand(s, args, ui)
		case "status":
			cmdErr = statusCommand(s, args, ui)
		case "delete":
			cmdErr = deleteCommand(s, args, ui)
		case "list":
			cmdErr = listCommand(s, args, ui)
		case "clipboard":
			cmdErr = clipboardCommand(s, args, ui)
		case "decrypt":
			cmdErr = decryptCommand(s, args, ui)
		case "reencrypt":
			cmdErr = reencryptCommand(s, args, ui)
		case "rotate":
			cmdErr = rotateCommand(s, args, ui)
		}

		// 1. If help was requested (-h), always prioritize it and return success.
		if errors.Is(cmdErr, flag.ErrHelp) {
			return nil
		}

		// 2. If it wasn't a help request and setup failed, report the setup error.
		if setupErr != nil {
			return fmt.Errorf("unable to setup environment configuration: %w", setupErr)
		}

		// 3. Otherwise, return the subcommand's actual error (if any).
		return cmdErr
	}

	return fmt.Errorf("unknown command: %s", cmd)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "privage: %v\n", err)
	os.Exit(1)
}

func parseCatArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("cat", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s cat [label]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Print the full contents of an encrypted file to stdout.\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  label  The label of the file to show\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return "", err
	}

	catArgs := fs.Args()
	if len(catArgs) == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", errors.New("cat command needs one argument (label)")
	}

	return catArgs[0], nil
}

func parseInitArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var slot string
	fs.StringVar(&slot, "piv-slot", "", "Use the yubikey slot key to encrypt the age private key")
	fs.StringVar(&slot, "p", "", "alias for -piv-slot")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s init [options]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", parseErr
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", parseErr)
		fs.Usage()
		return "", parseErr
	}

	return slot, nil
}

func parseAddArgs(args []string, ui UI) (string, string, error) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s add [category] [label]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Add a new encrypted file.\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  category  A category (e.g., 'credential' or any custom string)\n")
		fmt.Fprintf(fs.Output(), "  label     A label for credentials, or an existing file path\n")
	}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", "", parseErr
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", parseErr)
		fs.Usage()
		return "", "", parseErr
	}

	addArgs := fs.Args()
	if len(addArgs) != 2 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", "", errors.New("add command needs two arguments: <category> <label>")
	}

	cat := addArgs[0]
	if len(cat) > 32 {
		return "", "", errors.New("first argument (category) length is greater than max allowed")
	}

	label := addArgs[1]
	if len(label) > 128 {
		return "", "", errors.New("second argument (label) length is greater than max allowed")
	}

	return cat, label, nil
}

func parseShowArgs(args []string, ui UI) (string, string, error) {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s show [label] [field]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Show the contents of an encrypted file (formatted if it's a credential).\n")
		fmt.Fprintf(fs.Output(), "  If a field name is provided, only that field's value is printed.\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  label  The label of the file to show\n")
		fmt.Fprintf(fs.Output(), "  field  Optional: specific TOML field to show (e.g., api_key)\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", "", err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return "", "", err
	}

	showArgs := fs.Args()
	if len(showArgs) == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", "", errors.New("show command needs at least one argument (label)")
	}

	label := showArgs[0]
	var fieldName string
	if len(showArgs) > 1 {
		fieldName = showArgs[1]
	}

	return label, fieldName, nil
}