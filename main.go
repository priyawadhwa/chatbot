package main

import (
	"fmt"
	"github.com/priyawadhwa/chatbot/pkg/github"
	"os"
)

func main() {
	if err := execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute() error {
	client := github.NewGithubClient()
	cards, err := client.RetrieveCards()
	fmt.Println(cards)
	return err
}
