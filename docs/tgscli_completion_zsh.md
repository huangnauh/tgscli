## tgscli completion zsh

generate the autocompletion script for zsh

### Synopsis


Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions for every new session, execute once:
$ tgscli completion zsh > "${fpath[1]}/_tgscli"

You will need to start a new shell for this setup to take effect.


```
tgscli completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -f, --force         force
      --save-pinned   save pinned meta
      --verbose       verbose
```

### SEE ALSO

* [tgscli completion](tgscli_completion.md)	 - generate the autocompletion script for the specified shell

