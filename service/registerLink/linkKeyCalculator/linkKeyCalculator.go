package linkKeyCalculator

import (
	"math/rand"
	"strconv"
	"sync"
)

type LinkKeyCalculator struct {
	charMaps [][]string
	mux sync.RWMutex
	rng *rand.Rand
}

func New(seed int64) *LinkKeyCalculator {
	return &LinkKeyCalculator{
		charMaps: [][]string{},
		rng:      rand.New(rand.NewSource(seed)),
	}
}

func (calc *LinkKeyCalculator) shuffleChars() []string {
	var chars = [62]string{}
	for i := 0; i < 62; i++ {
		if i < 10 {
			chars[i] = strconv.Itoa(i)
		} else if i < 36 {
			chars[i] = string(i+55)
		} else {
			chars[i] = string(i+61)
		}
	}

	for i := 0; i < 62; i++ {
		r := calc.rng.Intn(62-i) + i
		t := chars[i]
		chars[i] = chars[r]
		chars[r] = t
	}

	return chars[:]
}

func (calc *LinkKeyCalculator) getCharMap(pos int) []string {
	calc.mux.RLock()
	if pos >= len(calc.charMaps) {
		calc.mux.RUnlock()
		calc.mux.Lock()
		for pos >= len(calc.charMaps) {
			calc.charMaps = append(calc.charMaps, calc.shuffleChars())
		}
		calc.mux.Unlock()
	} else {
		calc.mux.RUnlock()
	}

	return calc.charMaps[pos]
}

func (calc *LinkKeyCalculator) GetLinkKey(ordinal int) string {
	if ordinal == 0 {
		return calc.getCharMap(0)[0]
	}

	var num = ordinal
	var pos = 0
	var res = ""
	for num != 0 {
		q, rem := num/62, num%62
		num = q
		c := calc.getCharMap(pos)[rem]
		pos++
		res += c
	}

	return res
}