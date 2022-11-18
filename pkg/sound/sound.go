//go:build !libasound
// +build !libasound

// Copyright 2019 The Oto Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sound

import (
	"sync"
)

type SineWavePlayer struct{}

func NewSineWavePlayer(sampleRate int, channelCount int, format int) (*SineWavePlayer, chan struct{}, error) {
	sp := &SineWavePlayer{}
	ready := make(chan struct{}, 1)
	ready <- struct{}{}
	return sp, ready, nil
}

func (sp *SineWavePlayer) PlayLoop(wg *sync.WaitGroup) {
	wg.Done()
}

func (sp *SineWavePlayer) QueueNote(n Note) {}

func (sp *SineWavePlayer) Close() {}
