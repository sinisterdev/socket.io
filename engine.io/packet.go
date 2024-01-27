/**
 * Golang socket.io
 * Copyright (C) 2024 Kevin Z <zyxkad@gmail.com>
 * All rights reserved
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package engine

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/LiterMC/socket.io/internal/utils"
)

type UnexpectedPacketTypeError struct {
	Type PacketType
}

var _ error = (*UnexpectedPacketTypeError)(nil)

func (e *UnexpectedPacketTypeError) Error() string {
	return fmt.Sprintf("Unexpected packet type %d", e.Type)
}

type PacketType int8

const (
	unknownType PacketType = -1
	OPEN        PacketType = iota
	CLOSE
	PING
	PONG
	MESSAGE
	UPGRADE
	NOOP
)

func (t PacketType) String() string {
	switch t {
	case OPEN:
		return "OPEN"
	case CLOSE:
		return "CLOSE"
	case PING:
		return "PING"
	case PONG:
		return "PONG"
	case MESSAGE:
		return "MESSAGE"
	case UPGRADE:
		return "UPGRADE"
	case NOOP:
		return "NOOP"
	}
	return fmt.Sprintf("PacketType(%d)", (int8)(t))
}

func (t PacketType) ID() byte {
	switch t {
	case OPEN:
		return '0'
	case CLOSE:
		return '1'
	case PING:
		return '2'
	case PONG:
		return '3'
	case MESSAGE:
		return '4'
	case UPGRADE:
		return '5'
	case NOOP:
		return '6'
	}
	panic(&UnexpectedPacketTypeError{t})
}

func pktTypeFromByte(id byte) PacketType {
	switch id {
	case '0':
		return OPEN
	case '1':
		return CLOSE
	case '2':
		return PING
	case '3':
		return PONG
	case '4':
		return MESSAGE
	case '5':
		return UPGRADE
	case '6':
		return NOOP
	}
	return unknownType
}

type Packet struct {
	typ  PacketType
	body []byte
}

func (p *Packet) Type() PacketType {
	return p.typ
}

func (p *Packet) String() string {
	return fmt.Sprintf("Packet(%s, <%d bytes>)", p.typ.String(), len(p.body))
}

func (p *Packet) Body() []byte {
	return p.body
}

func (p *Packet) SetBody(body []byte) {
	p.body = body
}

func (p *Packet) UnmarshalBody(ptr any) error {
	return json.Unmarshal(p.body, &ptr)
}

func (p *Packet) WriteTo(w io.Writer) (n int64, err error) {
	if err = utils.WriteByte(w, p.typ.ID()); err != nil {
		return
	}
	n++
	var n0 int
	n0, err = w.Write(p.body)
	n += (int64)(n0)
	return
}

func (p *Packet) ReadFrom(r io.Reader) (n int64, err error) {
	var b byte
	b, err = utils.ReadByte(r)
	if err != nil {
		return
	}
	n++

	typ := pktTypeFromByte(b)
	if typ == unknownType {
		err = &UnexpectedPacketTypeError{typ}
		return
	}
	p.typ = typ
	// reuse buffer if possible
	p.body, err = utils.ReadAllTo(r, p.body[:0])
	n += (int64)(len(p.body))
	return
}
