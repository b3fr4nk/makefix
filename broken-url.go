package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

type Repo struct {
	repoName string
	repoURL string
	errors []string
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
			// Download the content of the Markdown file
			mdContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, content.GetPath(), nil)
			if err != nil {
				fmt.Printf("Error getting file content: %v\n", err)
				continue
			}

			// Decode content using GetContent() method
			contentStr, err := mdContent.GetContent()
			if err != nil {
				fmt.Printf("Error decoding file content: %v\n", err)
				continue
			}

			markdownContents = append(markdownContents, contentStr)
		}
	}
	return markdownContents, nil
}

func parseMarkdown(file string, repo *Repo){
	// search markdown file for links and check if they are working i.e. no http errors and not affiliated with makeschool

	// Define a regular expression to match links
	linkRegex := regexp.MustCompile(`\b(?:https?|ftp):\/\/[-A-Za-z0-9+&@#\/%?=~_|!:,.;]*[-A-Za-z0-9+&@#\/%=~_|]`)
	makeschoolRegex := regexp.MustCompile(`(?i)makeschool`)

	matches := makeschoolRegex.FindAllString(file, -1)

	if len(matches) > 0 {
		for _, match := range(matches) {
			repo.errors = append(repo.errors, match)
			repo.errorType = append(repo.errorType, "MAKESCHOOL")
		}
	}

	links := linkRegex.FindAllString(file, -1)
	for _, link := range(links) {
		c := colly.NewCollector()

		c.OnError(func(r *colly.Response, err error) {
			repo.errors = append(repo.errors, link)
			repo.errorType = append(repo.errorType, "URL")
		})

		c.Visit(link)
	}
}

func findErrors() []Repo{
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
			parseMarkdown(content, &tempRepo)
			if len(tempRepo.errors) > 0{
				urlErrors = append(urlErrors, tempRepo)
			}
		}

	}

	return urlErrors
}

func main() {


	err := godotenv.Load()
  	if err != nil {
		fmt.Print("error")
    	log.Fatal("Error loading .env file")
  	}

	errors := findErrors()

	fmt.Println(findErrors())
	
	errorsJSON, err := json.MarshalIndent(errors, "", " ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.WriteFile("results/repoErrors.json", errorsJSON, 0666)


}