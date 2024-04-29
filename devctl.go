package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/beclab/devbox/pkg/api/server"
	"github.com/beclab/devbox/pkg/development/command"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/store/db/model"

	"github.com/emicklei/go-restful/v3"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "devctl",
		Short: "devctl",
		Long:  ``,
	}

	createAppCmd := &cobra.Command{
		Use:   "createapp",
		Short: "createapp",
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := command.SetCreateConfigByPrompt()
			if err != nil {
				return
			}
			pwd, _ := os.Getwd()
			err = command.CreateApp().WithDir(pwd).Run(context.TODO(), cfg)
			if err != nil {
				fmt.Println("create app failed with error ", err)
				return
			}
			appData := model.DevApp{
				AppName: cfg.Name,
				DevEnv:  cfg.DevEnv,
				AppType: db.CommunityApp,
			}
			_, err = server.InsertDevApp(&appData)
			if err != nil {
				fmt.Printf("insert to db error: %v\n", err)
				return
			}
		},
	}

	installCmd := &cobra.Command{
		Use:   "install",
		Short: "install",
		Run: func(cmd *cobra.Command, args []string) {
			auth, err := cmd.Flags().GetString("auth")
			if err != nil {
				fmt.Printf("get auth argument failed: %v\n", err)
				return
			}
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				fmt.Printf("get name argument failed: %v\n", err)
				return
			}
			url := fmt.Sprintf("http://devbox-server.user-space-hysyeah:8080/api/command/install-app")
			client := resty.New()
			resp, err := client.R().
				SetHeader(restful.HEADER_ContentType, restful.MIME_JSON).
				SetCookie(&http.Cookie{Name: "auth_token", Value: auth}).
				SetBody(map[string]interface{}{"name": name, "source": "cli"}).
				Post(url)
			if err != nil {
				fmt.Printf("error occured: %v\n", err)
				return
			}
			if resp.StatusCode() != http.StatusOK {
				fmt.Printf("error occured with StatusCode: %d, error: %s, body: %s", resp.StatusCode(), resp.Error(), string(resp.Body()))
				return
			}
			fmt.Printf(resp.String())
		},
	}

	updateRepoCmd := &cobra.Command{
		Use:   "push [path]",
		Short: "push",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				fmt.Printf("need app path specified\n")
				return
			}
			path := args[0]
			_, err := cmd.Flags().GetBool("force")
			if err != nil {
				fmt.Printf("get force argument failed: %v\n", err)
				return
			}

			stat, err := os.Stat(path)
			if err != nil {
				fmt.Printf("stat error: %v", err)
				return
			}
			if os.IsNotExist(err) {
				fmt.Printf("path %s does not exist", path)
				return
			}
			if !stat.IsDir() {
				fmt.Printf("path is not a dir")
				return
			}
			dir := filepath.Join(path, "../")
			name := filepath.Base(path)
			err = command.UpdateRepo().WithDir(dir).Run(context.TODO(), name, true)
			if err != nil {
				fmt.Printf("push to charmuseum failed: %v\n", err)
			}
		},
	}

	installCmd.Flags().StringP("auth", "a", "", "auth token")
	installCmd.Flags().StringP("name", "n", "", "your application name")
	//installCmd.Flags().StringP("token", "t", "", "access token")

	updateRepoCmd.Flags().BoolP("force", "f", false, "force push even chart version existed")
	//updateRepoCmd.Flags().StringP("path", "p", "", "your application path")

	rootCmd.AddCommand(createAppCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(updateRepoCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
