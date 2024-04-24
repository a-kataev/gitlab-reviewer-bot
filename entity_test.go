package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueMergeRequestList(t *testing.T) {
	mrList := []*MergeRequest{
		{ProjectID: 1, ID: 3},
		{ProjectID: 2, ID: 1},
		{ProjectID: 2, ID: 2},
	}

	testCases := []struct {
		desc  string
		lists [][]*MergeRequest
		list  []*MergeRequest
	}{
		{
			desc: "",
			lists: [][]*MergeRequest{
				{mrList[0], mrList[0]},
				{mrList[0], mrList[0]},
			},
			list: []*MergeRequest{
				mrList[0],
			},
		},
		{
			desc: "",
			lists: [][]*MergeRequest{
				{mrList[1], mrList[2]},
				{mrList[0], mrList[1]},
			},
			list: []*MergeRequest{
				mrList[2], mrList[1], mrList[0],
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			list := UniqueMergeRequestList(tC.lists...)

			assert.Equal(t, list, tC.list)
		})
	}
}

func TestApprovalStatus(t *testing.T) {
	testCases := []struct {
		desc      string
		approvals []*User
		reviewers Reviewers
		status    ApprovalStatus
	}{
		{
			desc:      "",
			approvals: []*User{},
			reviewers: []int{1},
			status:    NoApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}},
			reviewers: []int{},
			status:    NoApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 4}},
			reviewers: []int{1},
			status:    UnknownApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 4}},
			reviewers: []int{1, 2, 3},
			status:    UnknownApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}},
			reviewers: []int{1},
			status:    SelfApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}},
			reviewers: []int{2},
			status:    SelfApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}},
			reviewers: []int{2, 4},
			status:    SelfApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}},
			reviewers: []int{2},
			status:    SelfApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 3}},
			reviewers: []int{4, 3},
			status:    OthersApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}, {ID: 3}},
			reviewers: []int{4, 3},
			status:    OthersApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}, {ID: 3}},
			reviewers: []int{2, 3, 1},
			status:    AllApprovals,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 4}},
			reviewers: []int{2, 3, 1},
			status:    AllApprovals,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := &MergeRequest{
				Approvals: tC.approvals,
			}

			status := mr.ApprovalStatus(tC.reviewers)
			if status != tC.status {
				t.Errorf("%v != %v", status, tC.status)
			}
		})
	}
}

func TestNextReviewer(t *testing.T) {
	testCases := []struct {
		desc      string
		approvals []*User
		reviewers Reviewers
		found     bool
		reviewer  int
	}{
		{
			desc:      "",
			approvals: []*User{},
			reviewers: []int{1},
			found:     true,
			reviewer:  1,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}},
			reviewers: []int{1},
			found:     true,
			reviewer:  0,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}, {ID: 3}},
			reviewers: []int{1},
			found:     true,
			reviewer:  0,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}, {ID: 3}},
			reviewers: []int{1, 3},
			found:     true,
			reviewer:  0,
		},
		{
			desc:      "",
			approvals: []*User{{ID: 1}, {ID: 2}},
			reviewers: []int{1, 2, 3},
			found:     true,
			reviewer:  3,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			mr := &MergeRequest{
				Approvals: tC.approvals,
			}

			found, reviewer := mr.NextReviewer(tC.reviewers)
			if found != tC.found {
				t.Errorf("%v != %v", found, tC.found)
			}
			if reviewer != tC.reviewer {
				t.Errorf("%v != %v", reviewer, tC.reviewer)
			}
		})
	}
}
