## tgscli completion bash

generate the autocompletion script for bash

### Synopsis


Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:
$ source <(tgscli completion bash)

To load completions for every new session, execute once:
Linux:
  $ tgscli completion bash > /etc/bash_completion.d/tgscli
MacOS:
  $ tgscli completion bash > /usr/local/etc/bash_completion.d/tgscli

You will need to start a new shell for this setup to take effect.
  

```
tgscli completion bash
```

### Options

```
  -h, --help   help for bash
```

### Options inherited from parent commands

```
  -f, --force         force
      --save-pinned   save pinned meta
  -v, --verbose       verbose
```

### SEE ALSO

* [tgscli completion](tgscli_completion.md)	 - generate the autocompletion script for the specified shell

