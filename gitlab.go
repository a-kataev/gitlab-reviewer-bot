package main

import (
	"context"
	"fmt"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Gitlab struct {
	client *gitlab.Client
}

func (g *Gitlab) CurrentUser(ctx context.Context) (*User, error) {
	value, _, err := g.client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       value.ID,
		Name:     value.Name,
		Username: value.Username,
	}, nil
}

func (g *Gitlab) GetUser(ctx context.Context, user int) (*User, error) {
	value, _, err := g.client.Users.GetUser(user, gitlab.GetUsersOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       value.ID,
		Name:     value.Name,
		Username: value.Username,
	}, nil
}

func (g *Gitlab) listMergeRequests(ctx context.Context, opts *gitlab.ListMergeRequestsOptions) ([]*MergeRequest, error) {
	list, _, err := g.client.MergeRequests.ListMergeRequests(opts, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	mergeRequests := make([]*MergeRequest, 0, len(list))

	for _, mr := range list {
		approvals, _, err := g.client.MergeRequests.GetMergeRequestApprovals(
			mr.ProjectID,
			mr.IID,
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, err
		}

		notes, _, err := g.client.Notes.ListMergeRequestNotes(
			mr.ProjectID,
			mr.IID,
			&gitlab.ListMergeRequestNotesOptions{},
			gitlab.WithContext(ctx),
		)
		if err != nil {
			return nil, err
		}

		countComment := 0
		for _, n := range notes {
			if !n.System && n.Type == gitlab.DiscussionNote {
				countComment++
			}
		}

		comments := make([]*Comment, 0, countComment)

		for _, n := range notes {
			if n.System && n.Type != gitlab.DiscussionNote {
				continue
			}

			comment := new(Comment)
			comment.ID = n.ID
			comment.Body = n.Body
			comment.Author = User{
				ID:       n.Author.ID,
				Name:     n.Author.Name,
				Username: n.Author.Username,
			}
			comment.CreatedAt = *n.CreatedAt

			comments = append(comments, comment)
		}

		mergeRequest := &MergeRequest{
			ID:        mr.IID,
			ProjectID: mr.ProjectID,
			Draft:     mr.Draft,
			Author: User{
				ID:       mr.Author.ID,
				Name:     mr.Author.Name,
				Username: mr.Author.Username,
			},
			Approvals: make([]*User, 0, len(approvals.ApprovedBy)),
			Branches:  fmt.Sprintf("%s to %s", mr.SourceBranch, mr.TargetBranch),
			Link:      mr.WebURL,
			Comments:  comments,
		}

		if len(mr.Reviewers) > 0 {
			mergeRequest.Reviewer = &User{
				ID:       mr.Reviewers[0].ID,
				Name:     mr.Reviewers[0].Name,
				Username: mr.Reviewers[0].Username,
			}
		}

		if mr.Assignee != nil {
			mergeRequest.Assignee = &User{
				ID:       mr.Assignee.ID,
				Name:     mr.Assignee.Name,
				Username: mr.Assignee.Username,
			}
		}

		for _, approverUser := range approvals.ApprovedBy {
			mergeRequest.Approvals = append(
				mergeRequest.Approvals,
				&User{
					ID:       approverUser.User.ID,
					Name:     approverUser.User.Name,
					Username: approverUser.User.Username,
				},
			)
		}

		mergeRequests = append(mergeRequests, mergeRequest)
	}

	return mergeRequests, nil
}

func (g *Gitlab) listMergeRequestsOptions() *gitlab.ListMergeRequestsOptions {
	return &gitlab.ListMergeRequestsOptions{
		State: gitlab.Ptr("opened"),
		Draft: gitlab.Ptr(false),
		WIP:   gitlab.Ptr("no"),
		Scope: gitlab.Ptr("all"),
	}
}

func (g *Gitlab) ListMergeRequestsByReviewer(ctx context.Context, reviewer int) ([]*MergeRequest, error) {
	opts := g.listMergeRequestsOptions()
	opts.ReviewerID = gitlab.ReviewerID(reviewer)

	return g.listMergeRequests(ctx, opts)
}

func (g *Gitlab) ListMergeRequestsByAuthor(ctx context.Context, author int) ([]*MergeRequest, error) {
	opts := g.listMergeRequestsOptions()
	opts.AuthorID = &author
	opts.ReviewerID = gitlab.ReviewerID(gitlab.UserIDNone)

	return g.listMergeRequests(ctx, opts)
}

func (g *Gitlab) ListMergeRequestsByAssignee(ctx context.Context, assignee int) ([]*MergeRequest, error) {
	opts := g.listMergeRequestsOptions()
	opts.AssigneeID = gitlab.AssigneeID(assignee)

	return g.listMergeRequests(ctx, opts)
}

func (g *Gitlab) UpdateMergeRequestReviewer(ctx context.Context, project int, mergeRequest int, reviewer int) error {
	_, _, err := g.client.MergeRequests.UpdateMergeRequest(
		project,
		mergeRequest,
		&gitlab.UpdateMergeRequestOptions{
			ReviewerIDs: &[]int{reviewer},
		},
		gitlab.WithContext(ctx),
	)

	return err
}
