// Copyright (c) 2022 The privage developers

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/revelaction/privage/setup"
)

var global setup.Options

var (
	BuildCommit string
	BuildTag    string
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global options] command [command options] [arguments...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  init       Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		fmt.Fprintf(os.Stderr, "  key        Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
		fmt.Fprintf(os.Stderr, "  status     Provide information about the current configuration.\n")
		fmt.Fprintf(os.Stderr, "  add        Add a new encrypted file.\n")
		fmt.Fprintf(os.Stderr, "  delete     Delete an encrypted file.\n")
		fmt.Fprintf(os.Stderr, "  list       list metadata of all/some encrypted files.\n")
		fmt.Fprintf(os.Stderr, "  show       Show the contents the an encripted file.\n")
		fmt.Fprintf(os.Stderr, "  cat        Print the full contents of an encrypted file to stdout.\n")
		fmt.Fprintf(os.Stderr, "  clipboard  Copy the credential password to the clipboard\n")
		fmt.Fprintf(os.Stderr, "  decrypt    Decrypt a file and write its content in a file named after the label\n")
		fmt.Fprintf(os.Stderr, "  reencrypt  Reencrypt all decrypted files that are already encrypted. (default is dry-run)\n")
		fmt.Fprintf(os.Stderr, "  rotate     Create a new age key and reencrypt every file with the new key\n")
		fmt.Fprintf(os.Stderr, "  bash       Dump bash complete script.\n")
		fmt.Fprintf(os.Stderr, "  help, h    Show help for a command.\n")
		fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
		fmt.Fprintf(os.Stderr, "  -h, --help\n")
		fmt.Fprintf(os.Stderr, "    	Show help for privage\n")
		fmt.Fprintf(os.Stderr, "  -c, -conf string\n")
		fmt.Fprintf(os.Stderr, "    	Use file as privage configuration file\n")
		fmt.Fprintf(os.Stderr, "  -k, -key string\n")
		fmt.Fprintf(os.Stderr, "    	Use file path for private key\n")
		fmt.Fprintf(os.Stderr, "  -p, -piv-slot string\n")
		fmt.Fprintf(os.Stderr, "    	The PIV slot for decryption of the age key\n")
		fmt.Fprintf(os.Stderr, "  -r, -repository string\n")
		fmt.Fprintf(os.Stderr, "    	Use file path as path for the repository\n")
		fmt.Fprintf(os.Stderr, "\nVersion: %s, commit %s\n", BuildTag, BuildCommit)
	}

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
	var err error
	switch cmd {
	case "init":
		err = initCommand(opts, args)
	case "key":
		err = keyCommand(opts, args)
	case "status":
		err = statusCommand(opts, args)
	case "add":
		err = addCommand(opts, args)
	case "delete":
		err = deleteCommand(opts, args)
	case "list":
		err = listCommand(opts, args)
	case "show":
		err = showCommand(opts, args)
	case "cat":
		err = catCommand(opts, args)
	case "clipboard":
		err = clipboardCommand(opts, args)
	case "decrypt":
		err = decryptCommand(opts, args)
	case "reencrypt":
		err = reencryptCommand(opts, args)
	case "rotate":
		err = rotateCommand(opts, args)
	case "bash":
		err = bashCommand(opts, args)
	case "complete":
		err = completeCommand(opts, args)
	case "help", "h":
		if len(args) > 0 {
			return runCommand(args[0], []string{"--help"}, opts)
		}
		flag.Usage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}

	return err
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "privage: %v\n", err)
	os.Exit(1)
}
