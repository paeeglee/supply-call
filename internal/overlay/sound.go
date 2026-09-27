package overlay

import (
	"encoding/binary"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	sampleRate = 44100
	beepHz     = 880
	beepMillis = 150
)

// beeper plays a short tone generated in code (no audio file).
type beeper struct {
	player *audio.Player
}

func newBeeper() *beeper {
	ctx := audio.NewContext(sampleRate)
	return &beeper{player: ctx.NewPlayerFromBytes(beepPCM())}
}

func (b *beeper) play() {
	b.player.Rewind()
	b.player.Play()
}

// beepPCM is a 16-bit stereo sine with a short fade in and out.
func beepPCM() []byte {
	n := sampleRate * beepMillis / 1000
	fade := n / 10
	buf := make([]byte, n*4)
	for i := 0; i < n; i++ {
		amp := 0.3
		if i < fade {
			amp *= float64(i) / float64(fade)
		} else if i > n-fade {
			amp *= float64(n-i) / float64(fade)
		}
		v := int16(amp * math.MaxInt16 * math.Sin(2*math.Pi*beepHz*float64(i)/sampleRate))
		binary.LittleEndian.PutUint16(buf[i*4:], uint16(v))
		binary.LittleEndian.PutUint16(buf[i*4+2:], uint16(v))
	}
	return buf
}
