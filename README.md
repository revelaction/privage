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


# Usage

## Create a credentials file

In `privage`, credentials are structured text (.toml files), that can not only contain passwords, but any other
data associated with a website, like API keys, 2-factor backup codes, etc. 

To add a barebone credentials file (that you can later edit), use the
command `add`, specifying a `category` (for credential files it should be
always `credential`) and a `label` (any string that good describes the website, f.ex.  `somewebsite.com@loginname`). 
u can use 


    privage add credential somewebsite.com@loginname
    An encrypted file was saved for 📖 somewebsite.com@loginname  🔖credential

`privage` will generate a password, put the password (among other fields) in a
`.toml` file and encrypt that file under the `category` 'credential'.

It is recommended to use some naming convention for the credentials label, like
`<url>@loginname`

You can now list the encrypted file with: 

    ls -al

    drwxrwxr-x  3 user user 4096 Sep 26 18:27 .
    drwxr-xr-x 29 user user 4096 Sep 25 21:43 ..
    -rw-rw-r--  1 user user  347 Sep 26 18:27 66ceb74807d0fd997566360b22ecbda1590ec35fbd3dd0ce88e15311a4e53faf.age
    drwxrwxr-x  7 user user 4096 Sep 26 18:16 .git
    -rw-------  1 user user    0 Sep 26 18:21 .gitignore
    -rw-------  1 user user  189 Sep 26 18:21 privage-key.txt

That long `age` file is the encrypted credential file.  The label
(somewebsite.com@loginname) and the category (credential) were encrypted along
with the credential information.

## Encrypt any file

`privage` can encrypt any file. You can use any category and label. 

For example, to encrypt the file `secret-plan.doc` under the category
`work`:


    privage add work secret-plan.doc

## List the encrypted files

To list the encrypted files, use `list`:

    privage list
    Found 2 total encrypted tracked files.

            📖 somewebsite.com@loginname  🔖credential
            💼 secret-plan.doc 🔖work


To list only encrypted files corresponding to the category `credential`:

    privage list credential
    Found 1 files matching your category 'credential' of a total of 2 tracked files.

            📖 somewebsite.com@loginname  🔖credential

The `list` command accepts a string for matching the labels and categories:

    privage list somew
    Found 1 files with name matching 'somew':

        📖 somewebsite.com@loginname  🔖credential


## Copy the password to the clipboard 

The command `clipboard` copies the credential password to the clipboard 

    privage clipboard somewebsite.com@loginname 
    The password for `somewebsite.com@loginname` is in the clipboard

Use the flag `-d` (`--delete`) to empty the clipboard.

    privage clipboard -d 

## Show the contents of a credentials file

the command `show` presents in the terminal the login and the password:

    privage show somewebsite.com@loginname

        Login:👤 loginname
        Password:🔑 ad81h4b54*)(y73

To show all the credentials file contents, use the flag `-a`

    privage show -a somewebsite.com@loginname
    #
    login = "loginname"
    password = "ad8Q1hD4b54*)(y73"

    email = ""
    url = "somewebsite.com"

    # API keys
    api_key = ""
    api_secret = ""
    api_name = ""
    api_passphrase = ""
    verification_code = ""

    # two factor backup code
    two_factor_auth = ""

    # Other fields can be put in multiline
    remarks = '''
    - xxxx
    '''
# Design

The content of a `privage` encrypted file is the byte concatenation of two
`age` encrypted payloads:

The first encrypted payload (the header) contains the file name and a category
(plus a version of the header). This encrypted payload is padded to 512 bytes.


The second encrypted payload contains the file contents.

When listing the encrypted files, `privage` scans all encrypted files, retrieves the
encrypted header payload and decrypts it, presenting the header.

When writing the encrypted file, `privage` hashes the decrypted header and uses
the hash as name of the encrypted file. Encrypted `privage` file names look like this:

    425020f87e753ebe4dba67a872de04b7ce7350a63af9f74c1b7c4d633b41573c.age
    5e107b8e3b57411d5661d05e54f755408dd12c831a6b63e8033885c211da1317.age



