# Sensu CLI

TODO

## Completion

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
source <(sensu-cli completion bash)
```

You can now source your `~/.bash_profile` or launch a new terminal to utilize completion.

```sh
source ~/.bash_profile
```

### Installation (ZSH)

Add the following to your `~/.zshrc`:

```bash
source <(sensu-cli completion zsh)
```

You can now source your `~/.zshrc` or launch a new terminal to utilize completion.

```sh
source ~/.zshrc
```

### Usage

sensu-cli:
> $ sensu-cli <kbd>Tab</kbd>
```
check       configure   event       user
asset       completion  entity      handler
```

sensu-cli:
> $ sensu-cli check <kbd>Tab</kbd>
```
create  delete  import  list
```
