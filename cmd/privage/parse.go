package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func parseCatArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("cat", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s cat [label]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Print the full contents of an encrypted file to stdout.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label  The label of the file to show\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s init [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", parseErr
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, parseErr)
		fs.Usage()
		return "", parseErr
	}

	return slot, nil
}

func parseAddArgs(args []string, ui UI) (string, string, error) {
	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s add [category] [label]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Add a new encrypted file.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  category  A category (e.g., 'credential' or any custom string)\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label     A label for credentials, or an existing file path\n")
	}

	if parseErr := fs.Parse(args); parseErr != nil {
		if errors.Is(parseErr, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", "", parseErr
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, parseErr)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s show [label] [field]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Show the contents of an encrypted file (formatted if it's a credential).\n")
		_, _ = fmt.Fprintf(fs.Output(), "  If a field name is provided, only that field's value is printed.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label  The label of the file to show\n")
		_, _ = fmt.Fprintf(fs.Output(), "  field  Optional: specific TOML field to show (e.g., api_key)\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", "", err
		}
		fs.SetOutput(ui.Err)
		_, _ = fmt.Fprintf(ui.Err, "Error: %v\n", err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s delete [label]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Delete an encrypted file.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label  The label of the file to delete\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s key\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Decrypt the age private key with the PIV key defined in the .privage.conf file.\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return err
		}
		fs.SetOutput(ui.Err)
		_, _ = fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return err
	}
	return nil
}

func parseStatusArgs(args []string, ui UI) error {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s status\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Provide information about the current configuration.\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return err
		}
		fs.SetOutput(ui.Err)
		_, _ = fmt.Fprintf(ui.Err, "Error: %v\n", err)
		fs.Usage()
		return err
	}
	return nil
}

func parseListArgs(args []string, ui UI) (string, error) {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s list [filter]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  List metadata of all or some encrypted files.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  filter  Optional filter for category or label name\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s clipboard [options] [label]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Copy the credential password to the clipboard.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label  The label of the credential to copy\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s decrypt [label]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Decrypt a file and write its content in a file named after the label\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nArguments:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  label  The label of the file to decrypt\n")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s reencrypt [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Reencrypt all decrypted files that are already encrypted. (default is dry-run)\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return false, false, err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
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
		_, _ = fmt.Fprintf(fs.Output(), "Usage: %s rotate [options]\n", os.Args[0])
		_, _ = fmt.Fprintf(fs.Output(), "\nDescription:\n")
		_, _ = fmt.Fprintf(fs.Output(), "  Create a new age key and reencrypt every file with the new key.\n")
		_, _ = fmt.Fprintf(fs.Output(), "\nOptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.SetOutput(ui.Out)
			fs.Usage()
			return false, "", err
		}
		fs.SetOutput(ui.Err)
		FprintErr(ui.Err, err)
		fs.Usage()
		return false, "", err
	}
	return clean, slot, nil
}

