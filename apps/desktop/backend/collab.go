// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/Its-Satyajit/reqly/internal/collab"
)

// CollabList returns collaborators for the workspace.
func (s *AppService) CollabList() ([]collab.Collaborator, error) {
	root := s.root
	if root == "" {
		root = "."
	}
	ws, err := collab.Load(collab.DefaultPath(root))
	if err != nil {
		return nil, err
	}
	return ws.Collaborators, nil
}

// CollabAdd adds a collaborator.
func (s *AppService) CollabAdd(user, role string) error {
	root := s.root
	if root == "" {
		root = "."
	}
	path := collab.DefaultPath(root)
	ws, err := collab.Load(path)
	if err != nil {
		return err
	}
	if ws.Path == "" {
		ws.Path = root
	}
	if err := collab.AddCollaborator(&ws, user, role); err != nil {
		return err
	}
	return collab.Save(path, ws)
}

// CollabRemove removes a collaborator.
func (s *AppService) CollabRemove(user string) error {
	root := s.root
	if root == "" {
		root = "."
	}
	path := collab.DefaultPath(root)
	ws, err := collab.Load(path)
	if err != nil {
		return err
	}
	if err := collab.RemoveCollaborator(&ws, user); err != nil {
		return err
	}
	return collab.Save(path, ws)
}
