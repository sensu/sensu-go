package completion

import (
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

const (
	bashUsage = `
# Make sure bash completion is installed. If you use a current Linux
# in a non-minimal installation, bash completion should be available.
# On a Mac, install with:
brew install bash-completion

# Then add the following to your ~/.bash_profile
if [ -f $(brew --prefix)/etc/bash_completion ]; then
. $(brew --prefix)/etc/bash_completion
fi

# On Ubuntu, install with:
apt update && apt install bash-completion

# Then add the following to your ~/.bash_profile
if [ -f /usr/share/bash-completion/bash_completion ]; then
. /usr/share/bash-completion/bash_completion
elif [ -f /etc/bash_completion ]; then
. /etc/bash_completion
fi

# After bash-completions are available add the following to your ~/.bash_profile
source <(` + cli.SensuCmdName + ` completion bash)

# You can source your ~/.bash_profile or launch a new terminal to utilize completion.
source ~/.bash_profile
	`
)

func genBashCompletion(rootCmd *cobra.Command) error {
	stdout := rootCmd.OutOrStdout()
	return rootCmd.GenBashCompletion(stdout)
}
