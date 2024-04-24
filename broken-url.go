package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Repo struct {
	repoName string
	repoURL string
	urls []string
	errorType []string
	
}

func isMarkdownFile(filename string) bool{
	return len(filename) > 3 && filename[len(filename)-3:] == ".md"
}

func getMarkdown(repoURL string) ([]string, error){
	var markdownContents []string

	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	// source of the token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	// client for token
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Parse the URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		return markdownContents, err
	}

	// Extract owner and repository name from the path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) != 2 {
		fmt.Println("Invalid GitHub repository URL")
		return markdownContents, err
	}

	// GitHub repository information
	owner := pathParts[0]
	repo := pathParts[1]
	ref := "main" // or the branch/tag you want to access

	opt := &github.RepositoryContentGetOptions{
		Ref:ref,
	}

	_, contents, _,  err := client.Repositories.GetContents(ctx, owner, repo, "/", opt)
	if err != nil{
		fmt.Printf("Error getting repository contents: %v\n", err)
		return markdownContents, err
	}

	//filter out non .md files
	for _, content := range(contents) {
		if content.GetType() == "file" && isMarkdownFile(content.GetName()){
			output, err := content.GetContent()
			if err != nil{
				fmt.Print("Error retrieving markdown file contents: %v\n", err)
				return markdownContents, err
			}
			markdownContents = append(markdownContents, output)
		}
	}
	return markdownContents, nil
}

func parseMarkdown(file []byte, repo *Repo) bool{
	// search markdown file for links and check if they are working i.e. no http errors and not makeschool

	return false
}

func findErrors() {
	urlErrors := []Repo{}

	token := os.Getenv("GITHUB_TOKEN")
	ctx := context.Background()
	// source of the token
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	// client for token
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	opt := github.RepositoryListOptions {

	}

	repos, _, err := client.Repositories.List(ctx, "Tech-at-DU", &opt)

	if err != nil {
		println(err)
		os.Exit(1)
	}

	for _, repo := range(repos) {
		tempRepo := Repo{}
		tempRepo.repoName = *repo.FullName
		tempRepo.repoURL = repo.GetHTMLURL()
		mdContents, err := getMarkdown(repo.GetHTMLURL())
		if err != nil {
			println(err)
			os.Exit(1)
		}

		for _, content := range(mdContents) {
			isError := parseMarkdown([]byte(content), &tempRepo)
			if isError {
				urlErrors = append(urlErrors, tempRepo)
			}
		}

	}

}

func main() {
	err := godotenv.Load()
  	if err != nil {
		fmt.Print("error")
    	log.Fatal("Error loading .env file")
  	}

	findErrors()
}