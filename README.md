# Command Builder

Utility tool to execute scripts and provide documentation and autosuggestions out of the box.

Currently autosuggestion is working only for ZSH.

## Scripts
Scripts are declared inside files named _.commands.yaml_.
Scripts are loaded automatically from the folder the terminal session is currently in and inside the user home folder.

An example of the script can be seen in [`/.commands.yaml`](/.commands.yaml).

## Install

1. Download from the [Release page](https://github.com/rascio/command-builder/releases/) the binary for your platform.
2. Extract the `cb` executable file from the archive
3. Copy it into a folder that is part of your `PATH` (and reload the session).
4. Execute `cb zsh-autocomplete`, it will print a line to be added in your ZSH configuration (`~/.zshrc`)
5. Start a new session and try `cb help`, or `cb <TAB><TAB>` for autosuggestion