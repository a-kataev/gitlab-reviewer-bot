package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"path"
	"sync"
	"time"

	"github.com/spf13/pflag"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

func main() {
	gitlabHost := "https://gitlab.com"
	gitlabToken := ""
	reviewerIDs := []int{}

	flag := pflag.NewFlagSet(path.Base(os.Args[0]), pflag.ContinueOnError)

	flag.IntSliceVar(&reviewerIDs, "reviewer-ids", reviewerIDs, "")
	flag.StringVar(&gitlabHost, "gitlab-host", gitlabHost, "")
	flag.StringVar(&gitlabToken, "gitlab-token", gitlabToken, "")

	if err := flag.Parse(os.Args[1:]); err != nil {
		if !errors.Is(err, pflag.ErrHelp) {
			slog.Error("flag error", slog.String("error", err.Error()))
		}

		os.Exit(1)
	}

	if len(reviewerIDs) < 2 {
		slog.Error("incorrect reviewer list", slog.String("error", "must be greater than or equal to 2"))

		os.Exit(1)
	}

	client, err := gitlab.NewClient(
		gitlabToken,
		gitlab.WithBaseURL(gitlabHost+"/api/v4"),
	)
	if err != nil {
		slog.Error("client error", slog.String("error", err.Error()))

		os.Exit(1)
	}

	glab := &Gitlab{client: client}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	user, err := glab.CurrentUser(ctx)
	if err != nil {
		slog.Error("current user error", slog.String("error", err.Error()))

		os.Exit(1)
	}

	slog.Info("current user",
		slog.Int("id", user.ID),
		slog.String("username", user.Username),
		slog.String("name", user.Name),
	)

	for _, id := range reviewerIDs {
		if _, err := glab.GetUser(ctx, id); err != nil {
			slog.Error("get reviewer error",
				slog.Int("reviewer_id", id),
				slog.String("error", err.Error()),
			)

			os.Exit(1)
		}
	}

	slog.Info("use reviewers", slog.Any("reviewer_ids", reviewerIDs))

	reviewers := NewReviewers(
		reviewerIDs[0], // self
		reviewerIDs[1:]...,
	)

	store := &Store{
		statuses: sync.Map{},
	}

	Review(ctx, store, glab, reviewers)

	for {
		select {
		case <-time.Tick(30 * time.Second):
			Review(ctx, store, glab, reviewers)
		case <-ctx.Done():
			return
		}
	}
}
