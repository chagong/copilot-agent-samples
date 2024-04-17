package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v61/github"
)

func listPullRequests(c *gin.Context, token string, owner string, name string) []*github.PullRequest {
	client := github.NewClient(nil).WithAuthToken(token)

	// List pull requests where you are a requested reviewer
	opt := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 3},
		State:       "open",
	}
	pulls, _, err := client.PullRequests.List(c, owner, name, opt)
	if err != nil {
		fmt.Printf("Error listing pull requests: %v\n", err)
		return nil
	}
	return pulls
}

func getPullRequest(c *gin.Context, token string, owner string, name string, id int) (*string, error) {
	client := github.NewClient(nil).WithAuthToken(token)

	comments, _, err := client.PullRequests.ListComments(c, owner, name, id, nil)
	if err != nil {
		return nil, err
	}
	commentsStr, _ := json.Marshal(comments)

	commits, _, err := client.PullRequests.ListCommits(c, owner, name, id, nil)
	if err != nil {
		return nil, err
	}
	commitsStr, _ := json.Marshal(commits)

	files, _, err := client.PullRequests.ListFiles(c, owner, name, id, nil)
	if err != nil {
		return nil, err
	}
	filesStr, _ := json.Marshal(files)

	resp := fmt.Sprintf("Comments: %s. Commits: %s. Files: %s", commentsStr, commitsStr, filesStr)
	return &resp, nil
}

func approvePullRequest(c *gin.Context, token string, owner string, name string, id int) error {

	client := github.NewClient(nil).WithAuthToken(token)

	event := "APPROVE"
	body := "Looks good to me!"
	req := github.PullRequestReviewRequest{
		Event: &event,
		Body:  &body,
	}

	_, _, err := client.PullRequests.CreateReview(c, owner, name, id, &req)
	return err
}
