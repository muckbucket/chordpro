package chordpro

import (
	"io"

	"github.com/chris-skud/chordpro2/chordpro/parse"
	"github.com/chris-skud/chordpro2/chordpro/types"
)

type OutputProcessor interface {
	Process(sheetLines []types.SheetLine, w io.Writer) error
}

func NewProcessor(outputProcessor OutputProcessor) *Processor {
	return &Processor{
		outputProcessor: outputProcessor,
	}

}

type Processor struct {
	outputProcessor OutputProcessor
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func (p *Processor) Process(reader io.Reader, writer io.Writer) error {

	l := parse.NewLexer(reader)
	var tokens []types.Token2
	var offset int
	var currLine int
	for {
		pos, tok, lit := l.Lex()
		if tok == types.EOF {
			break
		}

		// adjust the position of the chord to exclude the opening square bracket
		if tok == types.Chord {

			// move currLine to next line and reset offset if processing new line
			if pos.Line > currLine {
				// move to the next line and reset chord count
				currLine = pos.Line
				offset = 0
			}

			// the offset is a sum of characters that includes "[" + "chord" + "]"
			pos.Column = pos.Column - 1 - offset // account for opening bracket

			// add this chord to the offset for the next iteration, including 2 spaces for open/closed brackets
			offset = offset + len(lit) + 2
		}

		tokens = append(tokens, types.Token2{Pos: types.Position{Line: pos.Line, Column: pos.Column}, Typ: tok, Literal: lit})
	}

	// convert tokens slice into typed rows of token slices,
	// including the creation of lyric-chord sets (a given lyric)
	// usually (but not always) has a row above it with chords

	// Seems possible to apply a pipeline-like pattern to this processing, but it's not yeat clear
	// whether the linear input-as-last-output really fits as any processing may need the original token list vs. a previously
	// mutated sheet list. Stuff like formatting directives ({soc}...{eoc)) are a wrinkle.

	lineCount := tokens[len(tokens)-1].Pos.Line + 1
	sheetLines := make([]types.SheetLine, lineCount)
	for _, token := range tokens {
		linePos := token.Pos.Line
		switch token.Typ {
		case types.Chord:
			sheetLines[linePos].Type = types.LyricChord
			sheetLines[linePos].LyricChordSet.Chords = append(
				sheetLines[linePos].LyricChordSet.Chords,
				token,
			)
		case types.Lyric, types.Space:
			sheetLines[linePos].Type = types.LyricChord
			sheetLines[linePos].LyricChordSet.Lyrics = append(
				sheetLines[linePos].LyricChordSet.Lyrics,
				token,
			)
		}
	}

	check(
		p.outputProcessor.Process(sheetLines, writer),
	)

	// lineCount := tokens[len(tokens)-1].Pos.Line
	// blocks := make([]types.ContentBlock, lineCount+1)
	// for _, token := range tokens {
	// 	blockPos := token.Pos.Line - 1 // adjust for 0 based slice
	// 	switch token.Typ {
	// 	case types.Lyric:
	// 		blocks[blockPos].Content += token.Literal
	// 	case types.Chord:
	// 		blocks[blockPos].Content += token.Literal
	// 	}

	// }

	// contentBlocks := []types.ContentBlock{}

	return nil
}
