//go:build libasound
// +build libasound

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
	"fmt"
	"io"
	"math"
	"sync"
	"time"

	"github.com/hajimehoshi/oto/v2"
)

func formatByteLength(format int) int {
	switch format {
	case FormatFloat32LE:
		return 4
	case FormatUnsignedInt8:
		return 1
	case FormatSignedInt16LE:
		return 2
	default:
		panic(fmt.Sprintf("unexpected format: %d", format))
	}
}

type sineWave struct {
	freq   float64
	length int64
	pos    int64

	playerChannelCount int
	playerFormat       int
	playerSampleRate   int

	remaining []byte
}

func (s *sineWave) Read(buf []byte) (int, error) {
	if len(s.remaining) > 0 {
		n := copy(buf, s.remaining)
		copy(s.remaining, s.remaining[n:])
		s.remaining = s.remaining[:len(s.remaining)-n]
		return n, nil
	}

	if s.pos == s.length {
		return 0, io.EOF
	}

	eof := false
	if s.pos+int64(len(buf)) > s.length {
		buf = buf[:s.length-s.pos]
		eof = true
	}

	var origBuf []byte
	if len(buf)%4 > 0 {
		origBuf = buf
		buf = make([]byte, len(origBuf)+4-len(origBuf)%4)
	}

	length := float64(s.playerSampleRate) / float64(s.freq)

	num := formatByteLength(s.playerFormat) * s.playerChannelCount
	p := s.pos / int64(num)
	switch s.playerFormat {
	case FormatFloat32LE:
		for i := 0; i < len(buf)/num; i++ {
			bs := math.Float32bits(float32(math.Sin(2*math.Pi*float64(p)/length) * 0.3))
			for ch := 0; ch < s.playerChannelCount; ch++ {
				buf[num*i+4*ch] = byte(bs)
				buf[num*i+1+4*ch] = byte(bs >> 8)
				buf[num*i+2+4*ch] = byte(bs >> 16)
				buf[num*i+3+4*ch] = byte(bs >> 24)
			}
			p++
		}
	case FormatUnsignedInt8:
		for i := 0; i < len(buf)/num; i++ {
			const max = 127
			b := int(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < s.playerChannelCount; ch++ {
				buf[num*i+ch] = byte(b + 128)
			}
			p++
		}
	case FormatSignedInt16LE:
		for i := 0; i < len(buf)/num; i++ {
			const max = 32767
			b := int16(math.Sin(2*math.Pi*float64(p)/length) * 0.3 * max)
			for ch := 0; ch < s.playerChannelCount; ch++ {
				buf[num*i+2*ch] = byte(b)
				buf[num*i+1+2*ch] = byte(b >> 8)
			}
			p++
		}
	}

	s.pos += int64(len(buf))

	n := len(buf)
	if origBuf != nil {
		n = copy(origBuf, buf)
		s.remaining = buf[n:]
	}

	if eof {
		return n, io.EOF
	}
	return n, nil
}

type SineWavePlayer struct {
	sampleRate   int
	channelCount int
	format       int
	ctx          *oto.Context
	notesQueue   chan Note
}

func (sp *SineWavePlayer) newSineWave(freq float64, duration time.Duration) *sineWave {
	l := int64(sp.channelCount) * int64(formatByteLength(sp.format)) * int64(sp.sampleRate) * int64(duration) / int64(time.Second)
	l = l / 4 * 4
	return &sineWave{
		freq:               freq,
		length:             l,
		playerChannelCount: sp.channelCount,
		playerFormat:       sp.format,
		playerSampleRate:   sp.sampleRate,
	}
}

func (sp *SineWavePlayer) play(note Note) {
	duration := time.Duration(note.Duration) * time.Millisecond
	p := sp.ctx.NewPlayer(sp.newSineWave(float64(note.Freq), duration))
	p.Play()
	time.Sleep(duration)
	p.Close()
}

func NewSineWavePlayer(sampleRate int, channelCount int, format int) (*SineWavePlayer, chan struct{}, error) {
	sp := &SineWavePlayer{
		sampleRate:   sampleRate,
		channelCount: channelCount,
		format:       format,
		notesQueue:   make(chan Note),
	}
	ctx, ready, err := oto.NewContext(sp.sampleRate, sp.channelCount, sp.format)
	if err != nil {
		return nil, nil, err
	}
	sp.ctx = ctx
	return sp, ready, nil
}

func (sp *SineWavePlayer) PlayLoop(wg *sync.WaitGroup) {
	defer wg.Done()

	for n := range sp.notesQueue {
		sp.play(n)
	}
}

func (sp *SineWavePlayer) QueueNote(n Note) {
	sp.notesQueue <- n
}

func (sp *SineWavePlayer) Close() {
	close(sp.notesQueue)
}
