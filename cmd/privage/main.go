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
	setupUsage()

	// Global -h/--help should go to Stdout
	flag.CommandLine.SetOutput(os.Stdout)

	cmd, args, global, err := parseMainArgs()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		// If no command was provided, show usage and exit with 1
		flag.CommandLine.SetOutput(os.Stderr)
		flag.Usage()
		os.Exit(1)
	}

	if err := runCommand(cmd, args, global); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fatal(err)
	}
}

func parseMainArgs() (string, []string, setup.Options, error) {
	var opts setup.Options

	flag.StringVar(&opts.ConfigFile, "conf", "", "Use file as privage configuration file")
	flag.StringVar(&opts.ConfigFile, "c", "", "alias for -conf")
	flag.StringVar(&opts.KeyFile, "key", "", "Use file path for private key")
	flag.StringVar(&opts.KeyFile, "k", "", "alias for -key")
	flag.StringVar(&opts.PivSlot, "piv-slot", "", "The PIV slot for decryption of the age key")
	flag.StringVar(&opts.PivSlot, "p", "", "alias for -piv-slot")
	flag.StringVar(&opts.RepoPath, "repository", "", "Use file path as path for the repository")
	flag.StringVar(&opts.RepoPath, "r", "", "alias for -repository")

	flag.Parse()

	if flag.NArg() == 0 {
		return "", nil, opts, errors.New("no command provided")
	}

	cmd := flag.Arg(0)
	args := flag.Args()[1:]
	return cmd, args, opts, nil
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
	case "help":
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

func parseDeleteArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("delete", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s delete [label]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Delete an encrypted file.\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  label  The label of the file to delete\n")
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

	deleteArgs := fs.Args()
	if len(deleteArgs) == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", errors.New("delete command needs one argument (label)")
	}

	return deleteArgs[0], nil
}

func parseKeyArgs(args []string, ui UI) error {
	fs := flag.NewFlagSet("key", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s key\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return err
	}
	return nil
}

func parseStatusArgs(args []string, ui UI) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s status\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Provide information about the current configuration.\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return err
	}
	return nil
}

func parseListArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s list [filter]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  List metadata of all or some encrypted files.\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  filter  Optional filter for category or label name\n")
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

	listArgs := fs.Args()
	filter := ""
	if len(listArgs) > 0 {
		filter = listArgs[0]
	}
	return filter, nil
}

func parseClipboardArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("clipboard", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s clipboard [options] [label]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Copy the credential password to the clipboard.\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  label  The label of the credential to copy\n")
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

	clipArgs := fs.Args()
	if len(clipArgs) == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", errors.New("clipboard command needs one argument (label)")
	}
	return clipArgs[0], nil
}

func parseDecryptArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("decrypt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s decrypt [label]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Decrypt a file and write its content in a file named after the label\n")
		fmt.Fprintf(fs.Output(), "\nArguments:\n")
		fmt.Fprintf(fs.Output(), "  label  The label of the file to decrypt\n")
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

	decArgs := fs.Args()
	if len(decArgs) == 0 {
		fs.SetOutput(ui.Err)
		fs.Usage()
		return "", errors.New("decrypt command needs one argument (label)")
	}
	return decArgs[0], nil
}

func parseReencryptArgs(args []string, ui UI) (bool, bool, error) {
	fs := flag.NewFlagSet("reencrypt", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var force, clean bool
	fs.BoolVar(&force, "force", false, "Force encryption of the files.")
	fs.BoolVar(&force, "f", false, "alias for -force")
	fs.BoolVar(&clean, "clean", false, "Force encryption the files and also delete/clean the decrypted files.")
	fs.BoolVar(&clean, "c", false, "alias for -clean")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s reencrypt [options]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Reencrypt all decrypted files that are already encrypted. (default is dry-run)\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return false, false, err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return false, false, err
	}
	return force, clean, nil
}

func parseRotateArgs(args []string, ui UI) (bool, string, error) {
	fs := flag.NewFlagSet("rotate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	var clean bool
	var slot string
	fs.BoolVar(&clean, "clean", false, "Delete old Key's encrypted files. Rename new encrypted files and the new key")
	fs.BoolVar(&clean, "c", false, "alias for -clean")
	fs.StringVar(&slot, "piv-slot", "", "Use the yubikey slot to encrypt the age private key with the RSA Key")
	fs.StringVar(&slot, "p", "", "alias for -piv-slot")
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage: %s rotate [options]\n", os.Args[0])
		fmt.Fprintf(fs.Output(), "\nDescription:\n")
		fmt.Fprintf(fs.Output(), "  Create a new age key and reencrypt every file with the new key.\n")
		fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return false, "", err
		}
		fs.SetOutput(ui.Err)
		fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return false, "", err
	}
	return clean, slot, nil
}

func setupUsage() {
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
		fmt.Fprintf(output, "  help       Show help for a command.\n")
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
}