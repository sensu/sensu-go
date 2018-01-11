# Sensu CLI

`sensuctl` is the command line interface for managing your Sensu cluster.

## Auto-Completion

### Installation (Bash Shell)

Make sure bash completion is installed. If you use a current Linux
in a non-minimal installation, bash completion should be available.
On a Mac, install with:

```sh
brew install bash-completion
```

Then add the following to your `~/.bash_profile`:

```bash
if [ -f $(brew --prefix)/etc/bash_completion ]; then
. $(brew --prefix)/etc/bash_completion
fi
```

When bash-completion is available we can add the following to your `~/.bash_profile`:

```bash
source <(sensuctl completion bash)
```

You can now source your `~/.bash_profile` or launch a new terminal to utilize completion.

```sh
source ~/.bash_profile
```

### Installation (ZSH)

Add the following to your `~/.zshrc`:

```bash
source <(sensuctl completion zsh)
```

You can now source your `~/.zshrc` or launch a new terminal to utilize completion.

```sh
source ~/.zshrc
```

### Usage

sensuctl:
> $ sensuctl <kbd>Tab</kbd>
```
check       configure   event       user
asset       completion  entity      handler
```

sensuctl:
> $ sensuctl check <kbd>Tab</kbd>
```
create  delete  import  list
```

## Contributing

[UX Guidelines](GUIDELINES.md)
