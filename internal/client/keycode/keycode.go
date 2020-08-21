package keycode

import (
	"fmt"
	"strings"
)

var keycode = map[string]int{}

func ParseShortcut(s string) ([]string, error) {
	initKeymap()
	var (
		tokens []string
		vals   []string
	)
	tokens = strings.Split(s, "+")
	for _, t := range tokens {
		if val := keycode[strings.ToUpper(t)]; val != 0 {
			vals = append(vals, strings.ToLower(t))
		} else {
			return vals, fmt.Errorf("unable to parse token %s", t)
		}
	}
	return vals, nil
}

func initKeymap() {
	keycode = map[string]int{
		"ESCAPE":    0x1B,
		"TAB":       0x09,
		"BACKTAB":   0x08,
		"BACKSPACE": 0x08,
		"RETURN":    0x0D,
		"ENTER":     0x0D,
		"DELETE":    0x2E,
		"SYSREQ":    0x0100000a,

		"LEFT":  0x25,
		"UP":    0x26,
		"RIGHT": 0x27,
		"DOWN":  0x28,

		"SHIFT":      0xA0,
		"CONTROL":    0xA2,
		"CTRL":       0xA2,
		"META":       0x5B,
		"WIN":        0x5B,
		"WINKEY":     0x5B,
		"ALT":        0x12,
		"CAPSLOCK":   0x14,
		"NUMLOCK":    0x90,
		"SCROLLLOCK": 0x91,

		"F1":  0x70,
		"F2":  0x71,
		"F3":  0x72,
		"F4":  0x73,
		"F5":  0x74,
		"F6":  0x75,
		"F7":  0x76,
		"F8":  0x77,
		"F9":  0x78,
		"F10": 0x79,
		"F11": 0x7A,
		"F12": 0x7B,

		"SPACE": 0x20,

		"0": 0x30,
		"1": 0x31,
		"2": 0x32,
		"3": 0x33,
		"4": 0x34,
		"5": 0x35,
		"6": 0x36,
		"7": 0x37,
		"8": 0x38,
		"9": 0x39,

		"A": 0x41,
		"B": 0x42,
		"C": 0x43,
		"D": 0x44,
		"E": 0x45,
		"F": 0x46,
		"G": 0x47,
		"H": 0x48,
		"I": 0x49,
		"J": 0x4a,
		"K": 0x4b,
		"L": 0x4c,
		"M": 0x4d,
		"N": 0x4e,
		"O": 0x4f,
		"P": 0x50,
		"Q": 0x51,
		"R": 0x52,
		"S": 0x53,
		"T": 0x54,
		"U": 0x55,
		"V": 0x56,
		"W": 0x57,
		"X": 0x58,
		"Y": 0x59,
		"Z": 0x5a,
	}

}
