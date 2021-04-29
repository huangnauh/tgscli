## tgscli completion fish

generate the autocompletion script for fish

### Synopsis


Generate the autocompletion script for the fish shell.

To load completions in your current shell session:
$ tgscli completion fish | source

To load completions for every new session, execute once:
$ tgscli completion fish > ~/.config/fish/completions/tgscli.fish

You will need to start a new shell for this setup to take effect.


```
tgscli completion fish [flags]
```

### Options

```
  -h, --help              help for fish
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
  -f, --force         force
      --save-pinned   save pinned meta
  -v, --verbose       verbose
```

### SEE ALSO

* [tgscli completion](tgscli_completion.md)	 - generate the autocompletion script for the specified shell

