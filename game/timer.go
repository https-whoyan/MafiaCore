package game

import (
	"math/rand"
	"time"

	myTime "github.com/https-whoyan/MafiaCore/time"
)

// Used to simulate.

func getRandomDuration() time.Duration {
	minMilliSecond := myTime.FakeVotingMinSeconds * 1000
	maxMilliSecond := myTime.FakeVotingMaxSeconds * 1000
	randMilliSecondDuration := rand.Intn(maxMilliSecond-minMilliSecond+1) + minMilliSecond
	return time.Duration(randMilliSecondDuration) * time.Millisecond
}

func compareMaps[K comparable, E comparable](map1, map2 map[K]E) bool {
	var keys = make(map[K]struct{})
	for letter := range map1 {
		keys[letter] = struct{}{}
	}
	for letter := range map2 {
		keys[letter] = struct{}{}
	}

	for key := range keys {
		val1, ok1 := map1[key]
		val2, ok2 := map2[key]
		if !ok1 || !ok2 {
			return false
		}
		if val1 != val2 {
			return false
		}
	}
	return true
}

func (g *Game) timer(duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		select {
		case <-ticker.C:
			g.timerDone <- struct{}{}
			return
		case <-g.timerStop:
			return
		}
	}()
}

func (g *Game) randomTimer() {
	duration := getRandomDuration()
	g.timer(duration)
}
