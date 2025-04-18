// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package httpjson

import (
	"go.uber.org/multierr"

	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/go-concert/unison"

	v2 "github.com/elastic/beats/v7/filebeat/input/v2"
	inputcursor "github.com/elastic/beats/v7/filebeat/input/v2/input-cursor"
	stateless "github.com/elastic/beats/v7/filebeat/input/v2/input-stateless"
	"github.com/elastic/beats/v7/libbeat/statestore"
	"github.com/elastic/elastic-agent-libs/logp"
)

// inputManager wraps one stateless input manager
// and one cursor input manager. It will create one or the other
// based on the config that is passed.
type InputManager struct {
	stateless *stateless.InputManager
	cursor    *inputcursor.InputManager
}

var _ v2.InputManager = InputManager{}

func NewInputManager(log *logp.Logger, store statestore.States) InputManager {
	sim := stateless.NewInputManager(statelessConfigure)
	return InputManager{
		stateless: &sim,
		cursor: &inputcursor.InputManager{
			Logger:     log,
			StateStore: store,
			Type:       inputName,
			Configure:  cursorConfigure,
		},
	}
}

// Init initializes both wrapped input managers.
func (m InputManager) Init(grp unison.Group) error {
	return multierr.Append(
		m.stateless.Init(grp),
		m.cursor.Init(grp),
	)
}

// Create creates a cursor input manager if the config has a date cursor set up,
// otherwise it creates a stateless input manager.
func (m InputManager) Create(cfg *conf.C) (v2.Input, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}
	if len(config.Cursor) == 0 {
		return m.stateless.Create(cfg)
	}
	return m.cursor.Create(cfg)
}
