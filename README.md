# privage

`privage` is a terminal based password manager and general file encryption tool that
relies on [age](https://age-encryption.org/v1) for encryption. Optionally it uses a 
[yubikey](https://developers.yubico.com/PIV/)) for encryption of the age key.

The main goal of `privage` is to have your secrets (credentials and other
files) securely backed up in untrusted 3-party repositories whitout revealing
any secret information (not even the file name) to those 3-party repositories.


**WARNING: The author is not a cryptographer, and the code has not been reviewed. Use at your own risk.**

# Use Case

You may want to use `privage` if:

- You want to have your encrypted credentials and other secrets files in a revision control system repository (ex: git)
- You want to have backups of this repository in untrusted 3 party services (github, gitlab, bitbucket)
- You do not want to leak any information (not even the name of the files) in
  case of a breach of those 3 party services, which you otherwise should always assume.
  `privage` guarantees not leaking information because it also encrypts the metadata of the files. 
- You want to have one encrypted file per credential or secret file.
- You trust the computer running `privage`. `privage` uses unencrypted `age`
  keys, following the reasoning
  [here](https://github.com/FiloSottile/age#passphrase-protected-key-files).
  `privage` supports yubikeys to encrypt the age secret key.  


# Features

- `privage` uses the golang [age API](https://github.com/FiloSottile/age/blob/main/age.go) for encryption of files.
- `privage` can use a yubikey (PIV smart card) to encrypt the age secret key. See [Yubikey](#markdown-header-yubikey)   
- `privage` uses `categories` to allow classification of the encrypted files. 
- Encrypted files do not reveal any metadata. `privage` encrypted files names are hashes of the file name and the category. See [design](#markdown-header-design)   
- `privage` encrypts any kind of file, not only credentials/passwords.
- `privage` can easily (with one command) change the secret key and reencode all the files with the new key. See [rotate](#markdown-header-rotate)   
- `privage` tries to be simple: it does not wrap `git` or your editor: Use git to control your
  repository and use your preferred editor to edit credentials files.
- Powerful command completion. All commands have completion. See [Bash Completion](#markdown-header-bash-completion)   

