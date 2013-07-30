// Copyright 2013 Google, Inc.  All rights reserved.
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

package term

import (
	"fmt"
)

type rect struct {
	x, y, width, height int
}

func (r rect) grow(dw, dh int) rect {
	return rect{
		x:      r.x - dw,
		y:      r.y - dh,
		width:  r.width + 2*dw,
		height: r.height + 2*dh,
	}
}

func (r rect) String() string {
	return fmt.Sprintf("[%dx%d@(%d,%d)]", r.width, r.height, r.x, r.y)
}

type Region struct {
	tty     *TTY
	content rect
	border  borderStyle
}

func (t *TTY) NewRegion(w, h, x, y int) *Region {
	if t.screen == nil {
		return nil
	}

	if x < 0 || y < 0 {
		return nil
	}

	return &Region{
		tty:     t,
		content: rect{x, y, w, h},
	}
}

func (r *Region) SetBorder(style borderStyle) {
	if r.border == nil {
		r.content = r.content.grow(-1, -1)
	}
	r.border = style
	if r.border == nil {
		r.content = r.content.grow(1, 1)
	}
}

func (r *Region) SetPos(x, y int) {
	if r.border != nil {
		x, y = x+1, y+1
	}
	r.content.x, r.content.y = x, y
}

func (r *Region) SetSize(width, height int) {
	if r.border != nil {
		width, height = width-2, height-2
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	r.content.width, r.content.height = width, height
}

func (r *Region) Draw() {
	rect := r.content
	border := r.border != nil
	if border {
		rect = rect.grow(1, 1)
	}

	line := make([]byte, rect.width)
	start, end := 0, rect.width
	if border {
		start, end = 1, rect.width-1
	}

	for row := 0; row < rect.height; row++ {
		if row < 2 || row == rect.height-1 {
			fill := byte(' ')
			if border {
				switch row {
				case 0:
					fill = r.border[borderHorizontal]
					line[0] = r.border[borderTopLeft]
					line[len(line)-1] = r.border[borderTopRight]
				default:
					line[0] = r.border[borderVertical]
					line[len(line)-1] = r.border[borderVertical]
				case rect.height - 1:
					fill = r.border[borderHorizontal]
					line[0] = r.border[borderBottomLeft]
					line[len(line)-1] = r.border[borderBottomRight]
				}
			}
			for col := start; col < end; col++ {
				line[col] = fill
			}
		}
		r.tty.SetCursor(rect.x, rect.y+row)
		r.tty.echo(line...)
	}

	r.tty.SetCursor(r.content.x, r.content.y)
}

func (t *TTY) Clear() {
	t.echo('\x1b', '[', '2', 'J')
}

// SetCursor Places the cursor at the given x,y position.
//
// Both x and y start at 0 and increase right and down.
func (t *TTY) SetCursor(x, y int) {
	if t.screen == nil {
		return
	}
	fmt.Fprintf(t.screen, "\x1b[%d;%dH", y+1, x+1)
}

type borderStyle []byte

var SimpleBorder = borderStyle{
	'-', '|', // Horizontal, Vertical
	',', '+', '.', // Top: Left, Tee, Right
	'+', '+', '+', // Left Tee, Center, Right Tee
	'`', '+', '\'', // Bottom: Left, Tee, Right
}

var FancyBorder = borderStyle{
	196, 179, // Horizontal, Vertical
	218, 194, 191, // Top: Left, Tee, Right
	195, 197, 180, // Left Tee, Center, Right Tee
	192, 193, 217, // Bottom: Left, Tee, Right
}

const (
	borderHorizontal = iota
	borderVertical
	borderTopLeft
	borderTopTee
	borderTopRight
	borderLeftTee
	borderCenter
	borderRightTee
	borderBottomLeft
	borderBottomTee
	borderBottomRight
	borderCount
)
