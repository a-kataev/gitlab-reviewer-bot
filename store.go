package main

import "sync"

type Item struct {
	Key    string
	Status ApprovalStatus
}

type Store struct {
	statuses sync.Map
}

func (s *Store) Store(item *Item) bool {
	save := false

	if v, ok := s.statuses.Load(item.Key); ok {
		if status, ok := v.(ApprovalStatus); ok {
			if item.Status != status {
				save = true
			}
		}
	} else {
		save = true
	}

	if save {
		if item.Status == AllApprovals {
			s.statuses.Delete(item.Key)
		} else {
			s.statuses.Store(item.Key, item.Status)
		}
	}

	return save
}
