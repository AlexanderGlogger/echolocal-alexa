//go:build !linux && (board_checkers || board_cronos || board_rook)

package ble

import (
	"errors"
	"time"
)

var errDevice = errors.New("ble: the radio is only on the device")

// Port is the controller's HCI node.
type Port struct{}

func OpenNode(string) (*Port, error) { return nil, errDevice }

func (p *Port) Flush() error { return errDevice }
func (p *Port) Close() error { return nil }

func (p *Port) Write([]byte) (int, error) { return 0, errDevice }
func (p *Port) Send([]byte) (int, error)  { return 0, errDevice }

func (p *Port) Ask([]byte, time.Duration) (event, error) { return event{}, errDevice }
func (p *Port) Read(time.Duration) (event, error)        { return event{}, errDevice }
func (p *Port) Raw(time.Duration) ([]byte, error)        { return nil, errDevice }
func (p *Port) Sniff(time.Duration) []string             { return nil }
func (p *Port) next(time.Duration) (packet, error)       { return packet{}, errDevice }
func (p *Port) Settle(time.Duration)                     {}
