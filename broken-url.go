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
	RepoName string
	RepoURL string
	Errors []MakeError
	
}

type MakeError struct {
	Error string
	ErrorType string
	ErrorLocation string

}

func isMarkdownFile(filename string) bool{
	return len(filename) > 3 && filename[len(filename)-3:] == ".md"
}

func getMarkdown(repoURL string) ([]string, []string, error){
	var markdownContents []string
	var filePaths []string

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
		return markdownContents, filePaths, err
	}

	// Extract owner and repository name from the path
	pathParts := strings.Split(strings.Trim(parsedURL.Path, "/"), "/")
	if len(pathParts) != 2 {
		fmt.Println("Invalid GitHub repository URL")
		return markdownContents, filePaths, err
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
		return markdownContents, filePaths, err
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
			filePaths = append(filePaths, content.GetPath())
		}
	}
	return markdownContents, filePaths, nil
}

func parseMarkdown(file string, repo *Repo, filepath string){
	// search markdown file for links and check if they are working i.e. no http errors and not affiliated with makeschool

	// Define a regular expression to match links
	linkRegex := regexp.MustCompile(`\b(?:https?|ftp):\/\/[-A-Za-z0-9+&@#\/%?=~_|!:,.;]*[-A-Za-z0-9+&@#\/%=~_|]`)
	makeschoolRegex := regexp.MustCompile(`(?i)makeschool`)

	matches := makeschoolRegex.FindAllString(file, -1)

	if len(matches) > 0 {
		for _, match := range(matches) {
			newError := MakeError{Error: match, ErrorType: "MAKESCHOOL", ErrorLocation: filepath}
			repo.Errors = append(repo.Errors, newError)
		}
	}

	links := linkRegex.FindAllString(file, -1)
	for _, link := range(links) {
		c := colly.NewCollector()

		c.OnError(func(r *colly.Response, err error) {
			newError := MakeError{Error: link, ErrorType: "URL", ErrorLocation: filepath}
			repo.Errors = append(repo.Errors, newError)
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
		tempRepo.RepoName = *repo.FullName
		tempRepo.RepoURL = repo.GetHTMLURL()
		mdContents, filepaths, err := getMarkdown(repo.GetHTMLURL())
		if err != nil {
			println(err)
			os.Exit(1)
		}

		for i, content := range(mdContents) {
			parseMarkdown(content, &tempRepo, filepaths[i])
			if len(tempRepo.Errors) > 0{
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
	
	errorsJSON, err := json.MarshalIndent(errors, "", " ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.WriteFile("results/repoErrors.json", errorsJSON, 0666)


}