package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a suite of apps",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			appName := args[0]

			// Call etcd client to get the suite of apps
			apps := getAppsFromEtcd(appName)

			// Display the apps to be deployed
			fmt.Println("Apps to be deployed:")
			for _, app := range apps {
				fmt.Println(app)
			}

			// Run deployment
			deploy(apps)

			// Fetch health checks for each app
			healthChecks := make(map[string]string)
			for _, app := range apps {
				healthCheck := getHealthCheckFromEtcd(app)
				healthChecks[app] = healthCheck
			}

			// Execute relevant health functions
			for app, healthCheck := range healthChecks {
				executeHealthFunction(app, healthCheck)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
