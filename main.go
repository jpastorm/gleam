package main

import (
	"flag"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	xgithub "github.com/jpastorm/gleam/github"
	"github.com/jpastorm/gleam/models"
	"github.com/jpastorm/gleam/storage"
	"os"
	"strings"
)

const (
	PUBLIC_USER  = "public-user"
	PRIVATE_USER = "private-user"
	ORGANIZATION = "organization"
)

var Repositories []string
var Persist bool

type github interface {
	CloneRepositories(repoURLs []string) error
	ListRepositories() ([]string, error)
	Setup(token, username string)
}

func main() {
	accessibleFlag := flag.Bool("a", false, "enable accessible mode")
	resetFlag := flag.Bool("r", false, "reset configuration")

	flag.Parse()

	configStorage := storage.NewConfig()
	githubStrategy := map[string]github{
		ORGANIZATION: xgithub.NewOrganization(),
		PUBLIC_USER:  xgithub.NewPublicUser(),
		PRIVATE_USER: xgithub.NewPrivateUser(),
	}

	var configs models.Config
	if !*resetFlag && configStorage.ConfigExists() {
		configsStorage, err := configStorage.LoadConfig()
		if err != nil {
			log.Error("Error loading configuration:", err)
		}

		configs = configsStorage
	} else {
		err := configStorage.DeleteConfigFile()
		if err != nil {
			log.Error("Error Deleting config file:", err)
		}
		configs.FirstTime = true
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Gleam!!").
				Description("@jpastorm | made with ❤ from Perú")),

		huh.NewGroup(
			huh.NewNote().
				Title("Previous Data").
				Description(func() string {
					var sb strings.Builder
					keyword := func(s string) string {
						return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
					}

					fmt.Fprintf(&sb,
						"%s",
						keyword(fmt.Sprintf("Username: %s\nUserType: %s", configs.Username, configs.UserType)),
					)
					return sb.String()
				}())).WithHide(!configStorage.ConfigExists()),

		huh.NewGroup(huh.NewSelect[string]().
			Options(huh.NewOptions(PUBLIC_USER, PRIVATE_USER, ORGANIZATION)...).
			Title("Choose user type").
			Description("").
			Validate(func(t string) error {
				if t == "" {
					return fmt.Errorf("no user type selected")
				}
				return nil
			}).
			Value(&configs.UserType)).
			WithHide(!configs.FirstTime),

		huh.NewGroup(
			huh.NewInput().
				Value(&configs.Username).
				Title("What's your username?").
				Placeholder("jpastorm").
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("empty username")
					}
					return nil
				})).
			WithHide(!configs.FirstTime),

		huh.NewGroup(
			huh.NewInput().
				Value(&configs.Token).
				Title("What's your token?").
				Placeholder("github token").
				Validate(func(s string) error {

					return nil
				}).
				Description("To download and list private repositories")).
			WithHideFunc(func() bool {
				return configs.UserType == PUBLIC_USER || configs.Token != ""
			}),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Would you like persist your config?").
				Value(&Persist).
				Affirmative("Yes!").
				Negative("No.")).
			WithHide(!configs.FirstTime),
	).WithAccessible(*accessibleFlag).Run()
	if err != nil {
		log.Error("Uh oh:", err)
		os.Exit(1)
	}

	strategy := githubStrategy[configs.UserType]
	strategy.Setup(configs.Token, configs.Username)

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Repositories").
				Description("choose at least 1 option.").
				Options(func() []huh.Option[string] {
					var huhOptions []huh.Option[string]

					if strategy == nil {
						return huhOptions
					}

					repositories, err := strategy.ListRepositories()
					if err != nil {
						log.Fatalf("error ListRepositories %s", err.Error())
					}

					for _, option := range repositories {
						huhOptions = append(huhOptions, huh.NewOption(option, option))
					}

					return huhOptions
				}()...).
				Validate(func(t []string) error {
					if len(t) <= 0 {
						return fmt.Errorf("at least one option is required")
					}
					return nil
				}).
				Value(&Repositories).
				Filterable(true),
		),
	).WithAccessible(*accessibleFlag).Run()
	if err != nil {
		log.Error("Uh oh:", err)
		os.Exit(1)
	}

	if Persist {
		err := configStorage.SaveConfig(configs.Token, configs.Username, configs.UserType)
		if err != nil {
			log.Error("Error saving configuration:", err)
			os.Exit(1)
		}
	}

	prepareBurger := func() {
		err := strategy.CloneRepositories(Repositories)
		if err != nil {
			log.Fatal(err)
		}
	}

	_ = spinner.New().Title("Downloading...").Accessible(*accessibleFlag).Action(prepareBurger).Run()

	var sb strings.Builder
	keyword := func(s string) string {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Render(s)
	}
	fmt.Fprintf(&sb,
		"%s\n\n%s",
		lipgloss.NewStyle().Bold(true).Render("REPOSITORIES DOWNLOADED"),
		keyword(EnumerateList(Repositories)),
	)

	fmt.Println(
		lipgloss.NewStyle().
			Width(70).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Render(sb.String()),
	)
}

func EnumerateList(array []string) string {
	enumerated := ""
	for i, item := range array {
		enumerated += fmt.Sprintf("%d. %s\n", i+1, item)
	}
	return enumerated
}
