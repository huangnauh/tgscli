## tgscli completion powershell

generate the autocompletion script for powershell

### Synopsis


Generate the autocompletion script for powershell.

To load completions in your current shell session:
PS C:\> tgscli completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
tgscli completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
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

