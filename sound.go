package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Les sons sont fabriqués par le programme (aucun fichier à fournir).

const sampleRate = 44100

type sfx int

const (
	sPop sfx = iota
	sDrop
	sCoin
	sTick
	sUnlock
	sClick
	sAngry
	sHire
	sBuy
	sBadge
	sPhone
	sSeed
	sCount
)

type soundBank struct {
	ctx   *audio.Context
	data  [sCount][]byte
	muted bool
	last  [sCount]int64
	frame int64
}

func encode(samples []float64) []byte {
	out := make([]byte, len(samples)*4)
	for i, s := range samples {
		s = math.Max(-1, math.Min(1, s))
		v := int16(s * 32000)
		out[i*4], out[i*4+1] = byte(v), byte(v>>8)
		out[i*4+2], out[i*4+3] = byte(v), byte(v>>8)
	}
	return out
}

func synth(dur float64, f func(t float64) float64) []byte {
	n := int(dur * sampleRate)
	s := make([]float64, n)
	for i := range s {
		s[i] = f(float64(i) / sampleRate)
	}
	return encode(s)
}

func tone(freq, t float64) float64 { return math.Sin(2 * math.Pi * freq * t) }

func newSoundBank() *soundBank {
	sb := &soundBank{ctx: audio.NewContext(sampleRate)}
	rng := rand.New(rand.NewSource(1))
	noise := func() float64 { return rng.Float64()*2 - 1 }
	// récolte : petit « pop »
	sb.data[sPop] = synth(0.09, func(t float64) float64 {
		return math.Exp(-t*45) * 0.25 * tone(500+1400*math.Exp(-t*40), t)
	})
	// dépôt : « toc » de bois
	sb.data[sDrop] = synth(0.07, func(t float64) float64 {
		return math.Exp(-t*60) * (0.2*tone(320, t) + 0.08*noise())
	})
	// argent : « kling »
	sb.data[sCoin] = synth(0.35, func(t float64) float64 {
		a := math.Exp(-t*14) * (tone(1975, t) + 0.5*tone(2637, t))
		if t > 0.06 {
			a += math.Exp(-(t-0.06)*12) * (tone(2637, t) + 0.4*tone(3951, t))
		}
		return a * 0.11
	})
	// argent versé dans une zone : « tic »
	sb.data[sTick] = synth(0.04, func(t float64) float64 {
		return math.Exp(-t*90) * 0.16 * tone(1500, t)
	})
	// nouveau bâtiment : fanfare
	notes := []float64{523.25, 659.25, 783.99, 1046.5, 1318.5}
	sb.data[sUnlock] = synth(1.2, func(t float64) float64 {
		i := int(t / 0.11)
		if i > 4 {
			i = 4
		}
		lt := t - float64(i)*0.11
		d := 0.11
		if i == 4 {
			d = 0.75
		}
		e := math.Max(0, 1-lt/d) * math.Min(1, lt/0.01)
		f := notes[i]
		return e * 0.2 * (tone(f, t) + 0.3*tone(2*f, t) + 0.12*tone(3*f, t))
	})
	sb.data[sClick] = synth(0.06, func(t float64) float64 {
		return math.Exp(-t*70) * 0.3 * tone(900, t)
	})
	// client mécontent : « bzz » grave
	sb.data[sAngry] = synth(0.3, func(t float64) float64 {
		sq := 1.0
		if math.Sin(2*math.Pi*140*t) < 0 {
			sq = -1
		}
		return math.Max(0, 1-t/0.3) * 0.08 * sq
	})
	// embauche : deux notes joyeuses
	sb.data[sHire] = synth(0.4, func(t float64) float64 {
		f := 784.0
		if t > 0.12 {
			f = 1175
		}
		return math.Exp(-math.Mod(t, 0.12)*10) * 0.18 * (tone(f, t) + 0.3*tone(2*f, t))
	})
	// achat : « tchiing » de caisse enregistreuse
	sb.data[sBuy] = synth(0.5, func(t float64) float64 {
		a := 0.0
		if t < 0.05 {
			a += 0.25 * noise() * (1 - t/0.05)
		}
		a += math.Exp(-t*7) * 0.15 * (tone(2093, t) + 0.6*tone(3136, t))
		return a
	})
	// badge : petit carillon brillant
	bn := []float64{1046.5, 1318.5, 1568, 2093}
	sb.data[sBadge] = synth(0.9, func(t float64) float64 {
		v := 0.0
		for i, f := range bn {
			st := float64(i) * 0.08
			if t > st {
				v += math.Exp(-(t-st)*6) * tone(f, t)
			}
		}
		return v * 0.12
	})
	// nouvelle commande : sonnerie de téléphone à l'ancienne, « drriing… drriing »
	// (le marteau frappe deux cloches très vite, deux fois)
	sb.data[sPhone] = synth(1.75, func(t float64) float64 {
		var lt float64
		switch {
		case t < 0.7:
			lt = t
		case t >= 0.95 && t < 1.65:
			lt = t - 0.95
		default:
			return 0
		}
		strike := math.Mod(lt, 1.0/22) // 22 coups par seconde, en alternant les deux cloches
		k := int(lt * 22)
		f := 1180.0
		if k%2 == 1 {
			f = 1420
		}
		bell := tone(f, t) + 0.5*tone(f*2.76, t) + 0.25*tone(f*5.4, t)
		env := math.Exp(-strike*55) * math.Min(1, lt/0.01) * math.Min(1, (0.7-lt)/0.05)
		hum := 0.25 * (tone(1180, t) + tone(1420, t)) // les cloches qui vibrent encore
		return 0.11 * (env*bell + hum*math.Min(1, (0.7-lt)/0.05))
	})
	// le joueur plante une graine : un tout petit « pof » de terre, très doux
	sb.data[sSeed] = synth(0.06, func(t float64) float64 {
		return math.Exp(-t*70) * 0.05 * (tone(220+180*math.Exp(-t*50), t) + 0.3*noise())
	})
	return sb
}

func (sb *soundBank) play(s sfx) {
	if sb == nil || sb.muted {
		return
	}
	p := sb.ctx.NewPlayerFromBytes(sb.data[s])
	switch s {
	case sPop, sDrop:
		p.SetVolume(0.5)
	}
	p.Play()
}
