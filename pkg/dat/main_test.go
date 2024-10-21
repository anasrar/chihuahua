package dat_test

import (
	"testing"

	"github.com/anasrar/chihuahua/pkg/dat"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	t.Run("pl00.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/pl00.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x26a), d.EntryTotal)
		assert.Equal(t, uint32(0x26a), uint32(len(d.Entries)))
	})

	t.Run("ema0.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/ema0.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x2), d.EntryTotal)
		assert.Equal(t, uint32(0x2), uint32(len(d.Entries)))
	})

	t.Run("ema4.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/ema4.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x63), d.EntryTotal)
		assert.Equal(t, uint32(0x63), uint32(len(d.Entries)))
	})

	t.Run("ema6.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/ema6.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x86), d.EntryTotal)
		assert.Equal(t, uint32(0x86), uint32(len(d.Entries)))
	})

	t.Run("r006.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/r006.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x21), d.EntryTotal)
		assert.Equal(t, uint32(0x21), uint32(len(d.Entries)))
	})

	t.Run("r006.dat: SCP", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPathWithOffsetSize(d, "../../samples/r006.dat", 480, 1022720); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x15), d.EntryTotal)
		assert.Equal(t, uint32(0x15), uint32(len(d.Entries)))
	})

	t.Run("r100.dat", func(t *testing.T) {
		d := dat.New()
		if err := dat.FromPath(d, "../../samples/r100.dat"); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x28), d.EntryTotal)
		assert.Equal(t, uint32(0x28), uint32(len(d.Entries)))
	})
}
