package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"go.uber.org/ratelimit"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type Result struct {
	Number           int       `json:"number"`
	Title            string    `json:"title"`
	ReferenceCount   int       `json:"referenceCount"`
	CreatedAt        time.Time `json:"createdAt"`
	LastReferencedAt time.Time `json:"lastReferencedAt"`
}

type ResultList struct {
	Items []*Result `json:"items"`
}

type ByReferenceCount []*Result

func (r ByReferenceCount) Len() int           { return len(r) }
func (r ByReferenceCount) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ByReferenceCount) Less(i, j int) bool { return r[i].ReferenceCount < r[j].ReferenceCount }

type Job struct {
	issue *github.Issue
	name  int
}

type Controller struct {
	issues []*github.Issue
	result []*Result
	jobs   chan Job

	limiter ratelimit.Limiter
	client  *github.Client
}

var defaultListOptions = github.ListOptions{
	PerPage: 100,
}

func (c *Controller) SortedResult() []*Result {
	result := c.result
	sort.Sort(ByReferenceCount(result))
	return result
}

func (c *Controller) JSONResult() []byte {
	list := ResultList{Items: []*Result{}}
	for _, r := range c.result {
		list.Items = append(list.Items, r)
	}
	result, _ := json.Marshal(list)
	return result
}

func (c *Controller) Run() error {
	config := envToMap()
	c.client = github.NewClient(oauth2.NewClient(oauth2.NoContext, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config["GITHUB_API_KEY"]})))
	options := &github.IssueListByRepoOptions{
		State:       "open",
		Labels:      []string{"kind/test-flake"},
		Sort:        "created",
		ListOptions: defaultListOptions,
	}

	c.issues = []*github.Issue{}

	for {
		issues, response, err := c.client.Issues.ListByRepo("openshift", "origin", options)
		if err != nil {
			return err
		}
		c.issues = append(c.issues, issues...)
		if response.NextPage == 0 {
			break
		}
		options.ListOptions.Page = response.NextPage
	}

	log.Printf("number of flakes: %d", len(c.issues))

	jobs := make(chan Job, len(c.issues))
	results := make(chan Result, len(c.issues))

	for w := 1; w <= 3; w++ {
		go c.worker(w, jobs, results)
	}

	c.result = []*Result{}
	c.limiter = ratelimit.New(3)

	for i, issue := range c.issues {
		jobs <- Job{name: i, issue: issue}
	}
	close(jobs)

	for _, issue := range c.issues {
		result, ok := <-results
		if !ok {
			return fmt.Errorf("return channel closed prematurely for issue %d", *issue.Number)
		}
		c.result = append(c.result, &result)
	}

	return nil
}

func (c *Controller) worker(id int, jobs <-chan Job, results chan<- Result) {
	defer func() {
		log.Printf("worker %d quit", id)
	}()
	prev := time.Now()
	for j := range jobs {
		now := c.limiter.Take()
		result := Result{
			Title:  *j.issue.Title,
			Number: *j.issue.Number,
		}
		log.Printf("#%d (job %d/%d): fetching timeline for #%d", id, j.name, len(c.issues), *j.issue.Number)
		timeline, _, err := c.client.Issues.ListIssueTimeline("openshift", "origin", result.Number, &defaultListOptions)
		if err != nil {
			log.Printf("#%d (job %d): error occured while fetching timeline for #%d: %v", id, j.name, result.Number, err)
		}
		for _, t := range timeline {
			if t.Source == nil || t.Event == nil {
				continue
			}
			// TODO: Check if cross-reference was PR or issue
			if *t.Event == "cross-referenced" {
				result.ReferenceCount += 1
				result.LastReferencedAt = *t.CreatedAt
			}
		}
		log.Printf("#%d (job %d/%d): recording results for #%d (%s)", id, j.name, len(c.issues), *j.issue.Number, now.Sub(prev))
		prev = now
		results <- result
	}
}

func envToMap() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.Split(e, "=")
		if len(parts) < 2 {
			continue
		}
		result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return result
}
