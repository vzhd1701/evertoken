# evertoken

Extract authentication token from Evernote installation and exb database.

## Installation

[**Download the latest release**](https://github.com/vzhd1701/evertoken/releases/latest) for your OS.

## Usage

```console
$ evertoken -h
evertoken - Extract authentication token from Evernote.
https://github.com/vzhd1701/evertoken

  Usage:
    evertoken [new|legacy|legacy-exb]

  Subcommands:
    new          Extract token from new Evernote app.
    legacy       Extract token from legacy Evernote app.
    legacy-exb   Extract token from EXB database file.

  Flags:
       --version   Displays the program version string.
    -h --help      Displays help with available flag, subcommand, and positional value parameters.

$ evertoken legacy-exb -h
legacy-exb - Extract token from EXB database file.

  Usage:
    legacy-exb [exb]

  Positional Variables:
    exb   EXB database file path. (Required)
  Flags:
       --version    Displays the program version string.
    -h --help       Displays help with available flag, subcommand, and positional value parameters.
    -p --password   Password to decrypt token data, numeric volume serial.
    -b --brute      Brute force password start number, use either this or password option. (default: -1)
```

## Example output

```console
$ evertoken new
C:\Users\User\AppData\Roaming\Evernote\secure-storage\authtoken_user_111111111
========================
User ID   111111111
Username  example123
Email     example@mail.com
Token     S=s999:U=fffffff:E=fffffffffff:C=fffffffffff:P=1dd:A=en-w32-xauth-new:V=2:H=ffffffffffffffffffffffffffffffff
Token Exp 2031-07-23 12:06:35

$ evertoken legacy
C:\Users\User\Evernote\Databases\example123.exb
========================
User ID   111111111
Username  example123
Email     example@mail.com
Token     S=s999:U=fffffff:E=fffffffffff:C=fffffffffff:P=1dd:A=en-w32-xauth-new:V=2:H=ffffffffffffffffffffffffffffffff
Token Exp 2031-07-23 12:06:35
```

## How it works

Evernote app uses a special authentication token to identify the user when the app communicates with the Evernote 
server. This token is issued when the user logs in into the app and has an expiration date of 10 years from the login
time or until the user will decide to sign out manually. The token is stored encrypted. **evertoken** allows to decrypt
& extract it from the Evernote app.

Evernote used different forms of storage & encryption of the token throughout its history. Here is a brief 
description of the differences between the versions:

### Evernote Legacy (v6.**) [Windows]

The token is stored inside the SQLite database file with `*.exb` extension located in 
`C:\Users\<Username>\Evernote\Databases\user_name.exb`. The token is encrypted using AES256 CBC encryption. The key 
is derived using the system drive's Volume Serial number. So the database can be decrypted only with the knowledge of 
Volume Serial from the machine it was created on. It can be brute-forced since volume serial is just a 32bit 
integer with the possible value range of 0 through 4294967295, but it takes quite a bit of time nonetheless (~400hr 
with i7-4790).

**evertoken** will automatically scan for `*.exb` files in known locations when run with `evertoken legacy` command. It 
also supports Yinxiang (印象笔记). It will get the system drive Volume Serial to decrypt the token data. If the database 
was created on another machine, you would have to extract Volume Serial from there to decrypt the token.

You can point it to a specific `*.exb` file with `evertoken legacy-exb <exb_file>` command. This command also provides
options to use a custom password with `-p` option, or try brute-forcing the password with `-b` option.

### Evernote Legacy (v7.**) [macOS]

The token is stored in a Keychain in a macOS-specific format alongside other user information like email and login. The 
token is not encrypted or scrambled in any way.

**evertoken** will extract the token from this version of Evernote if you will run `evertoken legacy` command. The 
system will prompt you for the password because **evertoken** will attempt to access protected storage.

### Evernote (v10.**) [Windows & macOS]

Token is stored as json encoded string located in 
`C:\Users\<Username>\AppData\Roaming\Evernote\secure-storage\authtoken_user_<user_id>` for Windows and in 
`~/Library/Application Support/Evernote/secure-storage/authtoken_user_<user_id>` for macOS. The token is 
encrypted using AES256 CBC encryption. The decryption key is stored in Windows Credentials for Windows and in 
Keychain for macOS.

**evertoken** will extract the token from this version of Evernote if you will run `evertoken new` command. The
system will prompt you for the password because **evertoken** will attempt to access protected storage.
