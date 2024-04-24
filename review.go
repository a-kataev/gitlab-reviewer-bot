package main

import (
	"context"
	"log/slog"
)

func Review(ctx context.Context, store *Store, glab *Gitlab, reviewers Reviewers) {
	listByAuthor, err := glab.ListMergeRequestsByAuthor(ctx, reviewers.Self())
	if err != nil {
		slog.Error("merge requests list (by author) error", slog.String("error", err.Error()))

		return
	}

	listByReviewer, err := glab.ListMergeRequestsByReviewer(ctx, reviewers.Self())
	if err != nil {
		slog.Error("merge requests list (by reviewer) error", slog.String("error", err.Error()))

		return
	}

	listByAssignee, err := glab.ListMergeRequestsByAssignee(ctx, reviewers.Self())
	if err != nil {
		slog.Error("merge requests list (by assignee) error", slog.String("error", err.Error()))

		return
	}

	for _, mr := range UniqueMergeRequestList(listByAuthor, listByReviewer, listByAssignee) {
		log := slog.With()

		if mr.Assignee != nil {
			log = log.With(slog.Int("assignee_id", mr.Assignee.ID))
		}

		approvals := make([]int, 0, len(mr.Approvals))
		for _, approval := range mr.Approvals {
			approvals = append(approvals, approval.ID)
		}

		if len(approvals) > 0 {
			log = log.With(slog.Any("approval_ids", approvals))
		}

		if mr.Reviewer != nil {
			log = log.With(slog.Any("reviewer_id", mr.Reviewer.ID))
		}

		status := mr.ApprovalStatus(reviewers)

		log = log.With(
			slog.String("url", mr.Link),
			slog.String("status", string(status)),
		)

		item := &Item{
			Key:    mr.Key(),
			Status: status,
		}

		switch status {
		case SelfApprovals:
			fallthrough
		case AllApprovals:
			if ok, reviewer := mr.NextReviewer(reviewers); ok {
				action := false

				if reviewer == 0 {
					if mr.Reviewer == nil {
						continue
					}

					action = store.Store(item)
					if action {
						log.Info("unassign reviewer")
					}
				} else {
					log = log.With(
						slog.Any("new_reviewer_id", reviewer),
					)

					action = store.Store(item)
					if action {
						log.Info("assign new reviewer")
					}
				}

				if action {
					if err := glab.UpdateMergeRequestReviewer(
						ctx,
						mr.ProjectID,
						mr.ID,
						reviewer,
					); err != nil {
						log.Error("update merge requests error", slog.String("error", err.Error()))
					}
				}
			}
		case NoApprovals:
			fallthrough
		case UnknownApprovals:
			fallthrough
		case OthersApprovals:
			if action := store.Store(item); action {
				log.Info("")
			}
		}
	}
}
