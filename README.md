# Makefix

This utility is designed to help identify issues in the ACS course repositories by finding links that contain Makeschool content and or return a http error.

## How it works

Using the Github developer API to gather all markdown files in a repo the program then visits every link in each markdown file to confirm it is working and is not still linked to makeschool

## How to use it

### Prerequisites:

- Have go installed
- Clone the source code to a directory you can easily find and access

### Steps

1. Navigate to wherever you downloaded the source code
2. enter `go run broken-url.go` into your terminal
3. open the generated JSON and update the repositories

## What was used

Using the gogithub library, godotenv library, and the oauth2 library
