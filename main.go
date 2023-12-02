package main

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	// Find and load kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = home + "/.kube/config"
		} else {
			fmt.Println("Unable to find kubeconfig file")
			os.Exit(1)
		}
	}

	// Load the kubeconfig file to get contexts
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		fmt.Printf("Error loading kubeconfig: %s\n", err)
		os.Exit(1)
	}

	// Extract context names
	var contexts []string
	for context := range config.Contexts {
		contexts = append(contexts, context)
	}

	// Create prompt
	prompt := promptui.Select{
		Label: "Select Kubernetes Context",
		Items: contexts,
	}

	_, selectedContext, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	// Change the context
	config.CurrentContext = selectedContext
	if err := clientcmd.WriteToFile(*config, kubeconfig); err != nil {
		fmt.Printf("Error writing kubeconfig file: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Context switched to %q\n", selectedContext)
}
