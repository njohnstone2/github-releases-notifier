package main

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

func buildWatchList(logger log.Logger, releases, tags []string) []RepoConfig {
	repos := []RepoConfig{}
	for _, r := range releases {
		// check if repo already exists
		if contains(repos, r) {
			level.Debug(logger).Log("msg", "repo already exists in watchlist", "repo", r)
			continue
		}

		repository := RepoConfig{
			Name:     r,
			Releases: true,
		}
		level.Debug(logger).Log("msg", "appending release repo", "release", repository)
		repos = append(repos, repository)
	}

	for i, t := range tags {
		// check if repo already exists
		if contains(repos, t) {
			level.Debug(logger).Log("msg", "repo already exists in watchlist", "repo", t)

			repository := repos[i]
			repository.Tags = true
			repos[i] = repository
			level.Debug(logger).Log("msg", "updated repo", "repo", repos[i])

			continue
		}

		repository := RepoConfig{
			Name: t,
			Tags: true,
		}
		level.Debug(logger).Log("msg", "appending tag repo", "repo", repository)
		repos = append(repos, repository)
	}

	return repos
}

func contains(s []RepoConfig, r string) bool {
	for _, a := range s {
		if a.Name == r {
			return true
		}
	}
	return false
}
