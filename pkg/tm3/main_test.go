package tm3_test

import (
	"testing"

	"github.com/anasrar/chihuahua/pkg/tm3"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	t.Run("pl00.dat", func(t *testing.T) {
		d := tm3.New()
		if err := tm3.FromPathWithOffsetSize(d, "../../samples/pl00.dat", 4960, 149312); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0xd), d.EntryTotal)
		assert.Equal(t, uint32(0xd), uint32(len(d.Entries)))
	})

	t.Run("ema0.dat", func(t *testing.T) {
		d := tm3.New()
		if err := tm3.FromPathWithOffsetSize(d, "../../samples/ema0.dat", 32, 163904); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0xa), d.EntryTotal)
		assert.Equal(t, uint32(0xa), uint32(len(d.Entries)))
	})

	t.Run("ema4.dat", func(t *testing.T) {
		d := tm3.New()
		if err := tm3.FromPathWithOffsetSize(d, "../../samples/ema4.dat", 800, 60160); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x5), d.EntryTotal)
		assert.Equal(t, uint32(0x5), uint32(len(d.Entries)))
	})

	t.Run("ema6.dat", func(t *testing.T) {
		d := tm3.New()
		if err := tm3.FromPathWithOffsetSize(d, "../../samples/ema6.dat", 996672, 154432); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x18), d.EntryTotal)
		assert.Equal(t, uint32(0x18), uint32(len(d.Entries)))
	})

	t.Run("r100.dat: SCP", func(t *testing.T) {
		d := tm3.New()
		if err := tm3.FromPathWithOffsetSize(d, "../../samples/r100.dat", 800, 465152); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x45), d.EntryTotal)
		assert.Equal(t, uint32(0x45), uint32(len(d.Entries)))
	})
}
