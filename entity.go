package main

import (
	"fmt"
	"sort"
	"time"
)

type User struct {
	ID       int
	Name     string
	Username string
}

type Comment struct {
	ID        int
	Body      string
	Author    User
	CreatedAt time.Time
}

type MergeRequest struct {
	ID        int
	ProjectID int
	Draft     bool
	Author    User
	Assignee  *User
	Reviewer  *User
	Approvals []*User
	Branches  string
	Link      string
	Comments  []*Comment
}

type ApprovalStatus string

const (
	NoApprovals      ApprovalStatus = "NO"
	UnknownApprovals ApprovalStatus = "UNKNOWN"
	SelfApprovals                   = "SELF"
	AllApprovals                    = "ALL"
	OthersApprovals                 = "OTHERS"
)

func (mr *MergeRequest) ApprovalStatus(reviewers Reviewers) ApprovalStatus {
	if len(mr.Approvals) == 0 {
		return NoApprovals
	}

	if len(reviewers) == 0 {
		return NoApprovals
	}

	status := UnknownApprovals
	approve := len(reviewers)

	for _, approver := range mr.Approvals {
		for _, reviewer := range reviewers {
			if reviewer == approver.ID {
				approve--

				if approver.ID == reviewers.Self() {
					status = SelfApprovals
				} else {
					status = OthersApprovals
				}

				break
			}
		}
	}

	if status == OthersApprovals && approve == 0 {
		status = AllApprovals
	}

	return status
}

func (mr *MergeRequest) NextReviewer(reviewers Reviewers) (bool, int) {
	if len(reviewers) == 0 {
		return false, 0
	}

	if len(mr.Approvals) == 0 {
		return true, reviewers[0]
	}

	approve := len(reviewers)
	found := false

	for _, reviewer := range reviewers {
		found = false

		for _, approver := range mr.Approvals {
			if approver.ID == reviewer {
				found = true

				break
			}
		}

		if !found {
			return true, reviewer
		}

		approve--
	}

	if approve == 0 {
		return true, 0
	}

	return false, 0
}

func (mr *MergeRequest) Key() string {
	return fmt.Sprintf("%v-%v", mr.ProjectID, mr.ID)
}

func UniqueMergeRequestList(lists ...[]*MergeRequest) []*MergeRequest {
	cap := 0

	for _, l := range lists {
		cap += len(l)
	}

	list := make([]*MergeRequest, 0, cap)
	keys := make(map[string]struct{}, 0)

	for _, l := range lists {
		for _, mr := range l {
			key := mr.Key()

			if _, ok := keys[key]; !ok {
				keys[key] = struct{}{}

				list = append(list, mr)
			}
		}
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].ProjectID != list[j].ProjectID {
			return list[i].ProjectID > list[j].ProjectID
		}

		return list[i].ID > list[j].ID
	})

	return list
}

type Reviewers []int

func NewReviewers(self int, others ...int) Reviewers {
	rs := make([]int, 0, 1+len(others))
	rs = append(rs, self)

	if len(others) > 0 {
		rs = append(rs, others...)
	}

	return rs
}

func (rs Reviewers) Self() int {
	if len(rs) == 0 {
		return 0
	}

	return rs[0]
}
