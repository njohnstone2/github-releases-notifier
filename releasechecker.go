package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	githubql "github.com/shurcooL/githubql"
)

// Checker has a githubql client to run queries and also knows about
// the current repositories releases to compare against.
type Checker struct {
	logger   log.Logger
	client   *githubql.Client
	releases map[string]Repository
}

// Run the queries and comparisons for the given repositories in a given interval.
func (c *Checker) Run(interval time.Duration, repositories []RepoConfig, releases chan<- Repository) {
	if c.releases == nil {
		c.releases = make(map[string]Repository)
	}

	for {
		for _, repo := range repositories {
			s := strings.Split(repo.Name, "/")
			owner, name := s[0], s[1]

			nextRepo := Repository{}
			if repo.Releases {
				r, err := c.releaseQuery(owner, name)
				if err != nil {
					level.Warn(c.logger).Log(
						"msg", "failed to query the repository's releases",
						"owner", owner,
						"name", name,
						"err", err,
					)
					continue
				}

				nextRepo = r

				// For debugging uncomment this next line
				//releases <- nextRepo

				currRepo, ok := c.releases[repo.Name]

				// We've queried the repository for the first time.
				// Saving the current state to compare with the next iteration.
				if !ok {
					c.releases[repo.Name] = nextRepo
					continue
				}

				if nextRepo.Release.PublishedAt.After(currRepo.Release.PublishedAt) {
					releases <- nextRepo
					c.releases[repo.Name] = nextRepo
				} else {
					level.Debug(c.logger).Log(
						"msg", "no new release for repository",
						"owner", owner,
						"name", name,
					)
				}
			} else if repo.Tags {
				t, err := c.tagQuery(owner, name)
				if err != nil {
					level.Warn(c.logger).Log(
						"msg", "failed to query the repository's releases",
						"owner", owner,
						"name", name,
						"err", err,
					)
					continue
				}
				nextRepo = t

				// For debugging uncomment this next line
				//releases <- nextRepo

				currRepo, ok := c.releases[repo.Name]

				// We've queried the repository for the first time.
				// Saving the current state to compare with the next iteration.
				if !ok {
					c.releases[repo.Name] = nextRepo
					continue
				}

				if strings.Compare(nextRepo.Tag.Name, currRepo.Tag.Name) != 0 {
					releases <- nextRepo
					c.releases[repo.Name] = nextRepo
				} else {
					level.Debug(c.logger).Log(
						"msg", "no new tag for repository",
						"owner", owner,
						"name", name,
					)
				}
			}
		}
		time.Sleep(interval)
	}
}

// This should be improved in the future to make batch requests for all watched repositories at once
// TODO: https://github.com/shurcooL/githubql/issues/17

func (c *Checker) releaseQuery(owner, name string) (Repository, error) {
	var query struct {
		Repository struct {
			ID          githubql.ID
			Name        githubql.String
			Description githubql.String
			URL         githubql.URI

			Releases struct {
				Edges []struct {
					Node struct {
						ID          githubql.ID
						Name        githubql.String
						Description githubql.String
						URL         githubql.URI
						PublishedAt githubql.DateTime
					}
				}
			} `graphql:"releases(last: 1, orderBy: { field: CREATED_AT, direction: ASC})"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubql.String(owner),
		"name":  githubql.String(name),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.client.Query(ctx, &query, variables); err != nil {
		return Repository{}, err
	}

	repositoryID, ok := query.Repository.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert repository id to string: %v", query.Repository.ID)
	}

	if len(query.Repository.Releases.Edges) == 0 {
		return Repository{}, fmt.Errorf("can't find any releases for %s/%s", owner, name)
	}
	latestRelease := query.Repository.Releases.Edges[0].Node

	releaseID, ok := latestRelease.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert release id to string: %v", query.Repository.ID)
	}

	return Repository{
		ID:          repositoryID,
		Name:        string(query.Repository.Name),
		Owner:       owner,
		Description: string(query.Repository.Description),
		URL:         *query.Repository.URL.URL,

		Release: &Release{
			ID:          releaseID,
			Name:        string(latestRelease.Name),
			Description: string(latestRelease.Description),
			URL:         *latestRelease.URL.URL,
			PublishedAt: latestRelease.PublishedAt.Time,
		},
	}, nil
}

func (c *Checker) tagQuery(owner, name string) (Repository, error) {
	var query struct {
		Repository struct {
			ID          githubql.ID
			Name        githubql.String
			Description githubql.String
			URL         githubql.URI

			Refs struct {
				Nodes []struct {
					ID   githubql.ID
					Name githubql.String
				}
			} `graphql:"refs(refPrefix: \"refs/tags/\", first: 1, orderBy: { field: TAG_COMMIT_DATE, direction: DESC})"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubql.String(owner),
		"name":  githubql.String(name),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.client.Query(ctx, &query, variables); err != nil {
		return Repository{}, err
	}

	repositoryID, ok := query.Repository.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert repository id to string: %v", query.Repository.ID)
	}

	if len(query.Repository.Refs.Nodes) == 0 {
		return Repository{}, fmt.Errorf("can't find any tags for %s/%s", owner, name)
	}
	latestTag := query.Repository.Refs.Nodes[0]

	tagID, ok := latestTag.ID.(string)
	if !ok {
		return Repository{}, fmt.Errorf("can't convert tag id to string: %v", query.Repository.ID)
	}

	return Repository{
		ID:          repositoryID,
		Name:        string(query.Repository.Name),
		Owner:       owner,
		Description: string(query.Repository.Description),
		URL:         *query.Repository.URL.URL,

		Tag: &Tag{
			ID:   tagID,
			Name: string(latestTag.Name),
			// CommittedDate: latestTag.CommittedDate.Time,
		},
	}, nil
}
