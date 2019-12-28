package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNote_Snippet(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		assert := assert.New(t)

		note := Note{
			Body: `<p>test</p>`,
		}

		s, err := note.Snippet(3)
		assert.NoError(err)
		assert.Equal("<p>te…</p>", s)
	})

	t.Run("deeply nested", func(t *testing.T) {
		assert := assert.New(t)

		note := Note{
			Body: `<div><div><p>test</p></div></div>`,
		}

		s, err := note.Snippet(3)
		assert.NoError(err)
		assert.Equal("<div><div><p>te…</p></div></div>", s)
	})

	t.Run("with trailing tags", func(t *testing.T) {
		assert := assert.New(t)

		note := Note{
			Body: `<p>test<br/></p><br/>`,
		}

		s, err := note.Snippet(3)
		assert.NoError(err)
		assert.Equal("<p>te…</p>", s)
	})

	t.Run("fairly complex", func(t *testing.T) {
		assert := assert.New(t)

		note := Note{
			Body: `<h1>test</h1><div>a<strong>b</strong></div><blockquote>e</blockquote><div><br>f<br><br>g<br><br>h</div>`,
		}

		s, err := note.Snippet(6)
		assert.NoError(err)
		assert.Equal("<h1>test</h1><div>a<strong>b</strong></div>", s)
	})
}
