# Simple git client

## 0. Introduction

### 0.1 This project

This is a codecrafters project that implements a small git client in Go.

There is currently no staging, so add and commit are the same. It is not yet possible to pull a full repository. These are changes I intend to make. I also intend to break this file up into smaller files thereafter. 

I describe in the sections below the things that this .go program does do. There is a shell script yg1.sh that goes with this .go file and runs it. Of course, it can be built and run as standard, which is what the shell script automates. This .go file cannot be run correctly with "go run".

### 0.2 Git for our purposes

Future work will also expand on this readme document to describe some of the inner workings of how git stores objects (hence a skipped section). For now we consider git to have three types of objects: 

1/ blobs, which are general Zlib-compressed files; 

2/ trees, which are (compressed) files that contain lists of files and other directories; and 

3/ commits, which stores `meta-data' such as a timestamp, the author, committer and parent commits of information to be committed into a repository, along with arbitrary messages.

The standard git manual is here: https://git-scm.com/docs/user-manual

I found this [document](https://ftp.newartisans.com/pub/git.from.bottom.up.pdf) of John Wiegley's useful, but there are many "Git under the hood" resources available.

## 2. Manual 

### 2.1 `./yg1.sh init`

This command initialises a hidden `.git` directory, as well as `/objects` and `/HEAD` subdirectories.

### 2.2 `./yg1.sh cat-file -p <blob_sha>`

Git blob and tree objects are stored in the `.git/objects/` directory under `/xx/Z`. Here 'Z' is a 38 character string and 'xx' are two characters. Together, 'xxZ' is a 40 character string representing the 20-byte long SHA1 hash of the blob object in hexidecimal. The command `ls-tree` should be used to correctly read a tree.

This command opens the blob object with SHA1 hash `<blob_sha>`, decompresses it, and prints it.

The `-p` flag can be replaced with a `-t` flag, in which case only the type (blob or tree) of object will be printed out.


### 2.3 `./yg1.sh hash-object -w <filename>`

This writes a file into a blob object, computes its SHA1 hash and writes the object into `.git/objects`.


### 2.4 `./yg1.sh ls-tree [--name-only] <tree-hash>`

This lists the mode, type, SHA1 hash, and name of object listed in the tree.

This comes with a `--name-only` flag, whichl lists only the names of directories and files on the tree.

### 2.5 `./yg1.sh write-tree`

This writes the current directory recursively into a tree object.

### 2.6 `./yg1.sh commit-tree <tree_hash> [-p <parent_commit_hash>] -m <commit_message>`

This command creates a commit object for a tree with optional parents information (with a `-p` flag).


## 3. Licence

This code is distributed under the MIT licence.

