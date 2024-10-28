package scr_test

import (
	"testing"

	"github.com/anasrar/chihuahua/pkg/scr"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	t.Run("pl00.dat", func(t *testing.T) {
		s := scr.New()
		if err := scr.FromPathWithOffset(s, "../../samples/pl00.dat", 154272); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, uint32(0x1a), s.NodeTotal)
	})
}
