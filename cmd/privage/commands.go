package main

import (
	"github.com/urfave/cli/v2"
)

var allCommands = []*cli.Command{
	{
		Name:   "init",
		Action: initAction,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "piv-slot",
				Aliases: []string{"p"},
				Usage:   "Use the yubikey slot key to encrypt the age private key",
			},
		},
		Usage: "Add a .gitignore, age/yubikey key file to the current directory. Add a config file in the home directory.",
	},
	{
		Name:   "key",
		Action: keyAction,
		Usage:  "Decrypt the age private key with the PIV key defined in the .privage.conf file.",
	},
	{
		Name:   "status",
		Action: statusAction,
		Usage:  "Provide information about the current configuration.",
	},
	{
		Name:         "add",
		Action:       addAction,
		BashComplete: bashCompleteForAdd,
		Usage:        "Add a new encrypted file.",
	},
	{
		Name:         "delete",
		Action:       deleteAction,
		BashComplete: bashCompleteLabel,
		Usage:        "Delete an encrypted file.",
	},
	{
		Name:         "list",
		Action:       listAction,
		BashComplete: bashCompleteCategory,
		Usage:        "list metadata of all/some encrypted files.",
	},
	{
		Name:         "show",
		Action:       showAction,
		BashComplete: bashCompleteLabel,
		Usage:        "Show the contents the an encripted file.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all the contents of a file",
			},
		},
	},
	{
		Name:         "clipboard",
		Action:       clipboardAction,
		Usage:        "Copy the credential password to the clipboard",
		BashComplete: bashCompleteLabel,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete  the contents of the clipboard",
			},
		},
	},
	{
		Name:         "decrypt",
		Action:       decryptAction,
		Usage:        "Decrypt a file and write its content in a file named after the label",
		BashComplete: bashCompleteLabel,
	},
	{
		Name:   "reencrypt",
		Usage:  "Reencrypt all decrypted files that are already encrypted. (default is dry-run)",
		Action: reencryptAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force encryption of the files.",
			},
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "Force encryption the files and also delete/clean the decrypted files.",
			},
		},
	},
	{
		Name:   "rotate",
		Usage:  "Create a new age key and reencrypt every file with the new key",
		Action: rotateAction,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "clean",
				Aliases: []string{"c"},
				Usage:   "Delete old Key's encrypted files. Rename new encrypted files and the new key",
			},
			&cli.StringFlag{
				Name:    "piv-slot",
				Aliases: []string{"p"},
				Usage:   "Use the yubikey slot to encrypt the age private key with the RSA Key",
			},
		},
	},
	{
		Name:   "bash",
		Usage:  "Dump bash complete script.",
		Action: bashAction,
	},
}
