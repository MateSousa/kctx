package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

func main() {
	contexts, config, kubeconfig := LoadKubeConfig()

	helper := `
  Usage:
    kubectx [command]
  
  Available Commands:
    default Set the default Kubernetes context
    ls  List Kubernetes Contexts
    help  Help about any command

  Flags:
    -h, --help   help for kubectx commands

  Use "kubectx [command] --help" for more information about a command.`

	cmd := &cobra.Command{
		Use:   "kubectx",
		Short: "Switch Kubernetes Context",
		Run: func(cmd *cobra.Command, args []string) {
			ChangeContext(contexts, config, kubeconfig)
		},
	}

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List Kubernetes Contexts",
		Run: func(cmd *cobra.Command, args []string) {
			ListContexts(contexts, config.CurrentContext)
		},
	}

	cmd.AddCommand(lsCmd)
	cmd.SetHelpTemplate(helper)

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func LoadKubeConfig() (contexts []string, config *api.Config, kubeconfig string) {
	kubeconfigs := os.Getenv("KUBECONFIG")
	if kubeconfigs == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfigs = home + "/.kube/config"
		} else {
			fmt.Println("Unable to find kubeconfig file")
			os.Exit(1)
		}
	}
	config, err := clientcmd.LoadFromFile(kubeconfigs)
	if err != nil {
		fmt.Printf("Error loading kubeconfig: %s\n", err)
		os.Exit(1)
	}

	contexts = []string{}

	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	return contexts, config, kubeconfigs
}

func ChangeContext(contexts []string, config *api.Config, kubeconfig string) {
	ctxs := make([]string, len(contexts))

	green := color.New(color.FgGreen).SprintFunc()
	currentContextIndicator := green(" <- current context")

	for i, context := range contexts {
		if context == config.CurrentContext {
			ctxs[i] = context + currentContextIndicator
		} else {
			ctxs[i] = context
		}
	}

	prompt := promptui.Select{
		Label: "Select Kubernetes Context",
		Items: ctxs,
		Templates: &promptui.SelectTemplates{
			Active:   `🚀 {{ . | green }}`,
			Inactive: `{{ . }}`,
			Selected: `{{ "✔" | green }} {{ . }}`,
		},
	}

	_, selectedContext, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			fmt.Printf("Goodbye 👋\n")
			os.Exit(0)
		} else {
			fmt.Printf("Prompt selection failed: %v\n", err)
			os.Exit(1)
		}
		return
	}
	selectedContextWithoutIndicator := strings.TrimSuffix(selectedContext, currentContextIndicator)

	config.CurrentContext = selectedContextWithoutIndicator
	if err := clientcmd.WriteToFile(*config, kubeconfig); err != nil {
		fmt.Printf("Error writing kubeconfig file: %s\n", err)
		os.Exit(1)
	}

	green = color.New(color.FgGreen).SprintFunc()
	fmt.Printf(green("Switched to context: %s\n"), selectedContext)
	os.Exit(0)
}

func ListContexts(contexts []string, currentContext string) {
	green := color.New(color.FgGreen).SprintFunc()
	for _, context := range contexts {
		if context == currentContext {
			fmt.Println(context, green("<- current context"))
		} else {
			fmt.Println(context)
		}
	}
}
