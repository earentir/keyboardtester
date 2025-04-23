package main

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// Key represents a key on the keyboard
type Key struct {
	Label      string
	X, Y, W, H int
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("failed to create screen: %v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("failed to init screen: %v", err)
	}
	defer s.Fini()

	keys := initKeys()
	logs := []string{}
	pressed := map[string]bool{}
	escCount, enterCount, spaceCount := 0, 0, 0

	// initial draw
	drawAll(s, keys, logs, pressed)
	s.Show()

	for {
		ev := s.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			// --- exit logic ---
			switch ev.Key() {
			case tcell.KeyEscape:
				escCount++
				if escCount >= 5 {
					return
				}
			case tcell.KeyEnter:
				enterCount++
				if enterCount >= 5 {
					return
				}
			case tcell.KeyRune:
				if ev.Rune() == ' ' {
					spaceCount++
					if spaceCount >= 5 {
						return
					}
				}
			}

			// --- mark pressed keys permanently ---
			mainLabel := labelFromEvent(ev)
			pressed[mainLabel] = true
			if ev.Modifiers()&tcell.ModCtrl != 0 || (ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ) {
				pressed["Ctrl"] = true
			}
			if ev.Modifiers()&tcell.ModAlt != 0 {
				pressed["Alt"] = true
			}
			if ev.Modifiers()&tcell.ModShift != 0 {
				pressed["Shift"] = true
			}
			// CapsLock heuristic
			if ev.Key() == tcell.KeyRune {
				r := ev.Rune()
				if unicode.IsLetter(r) && unicode.IsUpper(r) && ev.Modifiers()&tcell.ModShift == 0 {
					pressed["CapsLock"] = true
				}
			}

			// --- append to log ---
			ts := time.Now().Format("15:04:05")
			code := int(ev.Key())
			mods := modString(ev.Modifiers())
			logs = append(logs, fmt.Sprintf("%s | %-7s | Code=%3d | Mods=%s", ts, mainLabel, code, mods))

			// --- safe trim ---
			_, scrH := s.Size()
			sepY := keys[len(keys)-1].Y + keys[len(keys)-1].H
			maxLines := scrH - sepY - 1

			if maxLines <= 0 {
				// no room at all
				logs = []string{}
			} else if len(logs) > maxLines {
				// only keep the bottom-most maxLines entries
				logs = logs[len(logs)-maxLines:]
			}

			// --- redraw & show ---
			drawAll(s, keys, logs, pressed)
			s.Show()

		case *tcell.EventResize:
			s.Sync()
		}
	}
}

func initKeys() []Key {
	var out []Key
	addRow := func(labels []string, y int) {
		x := 0
		for _, L := range labels {
			w := len(L) + 2
			out = append(out, Key{Label: L, X: x, Y: y, W: w, H: 3})
			x += w + 1
		}
	}
	addRow([]string{"Esc", "F1", "F2", "F3", "F4", "F5", "F6", "F7", "F8", "F9", "F10", "F11", "F12"}, 0)
	addRow([]string{"`", "1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "-", "=", "Backspace"}, 4)
	addRow([]string{"Tab", "Q", "W", "E", "R", "T", "Y", "U", "I", "O", "P", "[", "]", "\\"}, 8)
	addRow([]string{"CapsLock", "A", "S", "D", "F", "G", "H", "J", "K", "L", ";", "'", "Enter"}, 12)
	addRow([]string{"Shift", "Z", "X", "C", "V", "B", "N", "M", ",", ".", "/", "Shift"}, 16)
	addRow([]string{"Fn", "Ctrl", "Win", "Alt", "Space", "Alt", "Win", "Menu", "Ctrl"}, 20)
	addRow([]string{"Insert", "Home", "PgUp"}, 24)
	addRow([]string{"Delete", "End", "PgDn"}, 28)
	addRow([]string{"Left", "Down", "Right", "Up"}, 32)
	return out
}

func drawAll(s tcell.Screen, keys []Key, logs []string, pressed map[string]bool) {
	s.Clear()
	blue := tcell.StyleDefault.Background(tcell.ColorBlue)

	// draw keyboard
	for _, k := range keys {
		if pressed[k.Label] {
			drawKey(s, k, blue)
		} else {
			drawKey(s, k, tcell.StyleDefault)
		}
	}

	// separator line
	w, _ := s.Size()
	sepY := keys[len(keys)-1].Y + keys[len(keys)-1].H
	for x := 0; x < w; x++ {
		s.SetContent(x, sepY, '-', nil, tcell.StyleDefault)
	}

	// draw log lines
	for i, line := range logs {
		for j, r := range line {
			if j >= w {
				break
			}
			s.SetContent(j, sepY+1+i, r, nil, tcell.StyleDefault)
		}
	}
}

func drawKey(s tcell.Screen, k Key, style tcell.Style) {
	for dx := 0; dx < k.W; dx++ {
		for dy := 0; dy < k.H; dy++ {
			s.SetContent(k.X+dx, k.Y+dy, ' ', nil, style)
		}
	}
	start := k.X + (k.W-len(k.Label))/2
	for i, r := range k.Label {
		s.SetContent(start+i, k.Y, r, nil, style)
	}
}

func labelFromEvent(ev *tcell.EventKey) string {
	switch ev.Key() {
	case tcell.KeyEscape:
		return "Esc"
	case tcell.KeyEnter:
		return "Enter"
	case tcell.KeyTab:
		return "Tab"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "Backspace"
	case tcell.KeyF1, tcell.KeyF2, tcell.KeyF3, tcell.KeyF4,
		tcell.KeyF5, tcell.KeyF6, tcell.KeyF7, tcell.KeyF8,
		tcell.KeyF9, tcell.KeyF10, tcell.KeyF11, tcell.KeyF12,
		tcell.KeyHome, tcell.KeyEnd, tcell.KeyInsert, tcell.KeyDelete,
		tcell.KeyPgUp, tcell.KeyPgDn, tcell.KeyUp, tcell.KeyDown,
		tcell.KeyLeft, tcell.KeyRight:
		return tcell.KeyNames[ev.Key()]
	case tcell.KeyRune:
		if ev.Rune() == ' ' {
			return "Space"
		}
		return strings.ToUpper(string(ev.Rune()))
	default:
		if ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ {
			return string('A' + rune(ev.Key()-tcell.KeyCtrlA))
		}
		return fmt.Sprintf("Key[%d]", ev.Key())
	}
}

func modString(m tcell.ModMask) string {
	var parts []string
	if m&tcell.ModCtrl != 0 {
		parts = append(parts, "Ctrl")
	}
	if m&tcell.ModAlt != 0 {
		parts = append(parts, "Alt")
	}
	if m&tcell.ModShift != 0 {
		parts = append(parts, "Shift")
	}
	if len(parts) == 0 {
		return "None"
	}
	return strings.Join(parts, "|")
}
