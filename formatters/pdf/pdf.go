package pdf

import (
	"io"
	"strings"

	"github.com/chordpro-ensemble/pkg/types"
	"github.com/phpdave11/gofpdf"
)

type Processor struct {
	Config types.Config
}

// Processor is the PDF specific processing of the token stream (lexed song)
// whose bytes are written to the passed writer. It does need knowledge
// of the lexer grammar
func (p *Processor) Process(metaDirectives types.MetaDirectives, sheetLines []types.SheetLine, w io.Writer) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	title := metaDirectives.Title()

	// file metadata
	pdf.SetAuthor("Chris Skudlarczyk", false)
	pdf.SetTitle(title, true)

	// header
	pdf.SetFont("Helvetica", "", 12)
	wd := pdf.GetStringWidth(title) + 6
	pdf.SetX((210 - wd) / 2)
	pdf.Cell(wd, 9, title)
	pdf.Ln(10)

	var chorusEnvironment bool
	for _, line := range sheetLines {
		switch line.Type {
		case types.LineLyric:
			// Process the lyrics line
			pdf.SetX(5)
			for _, lyricToken := range line.LyricTokens {

				// ok, this is a hack around what may be a larger
				// issue with the pdf package not properly supporting utf-8 chars
				lyricToken.Literal = strings.ReplaceAll(lyricToken.Literal, "’", "'")

				preX, preY := pdf.GetXY()
				w := pdf.GetStringWidth(lyricToken.Literal)
				if chorusEnvironment {
					pdf.SetFont("Helvetica", "", 16)
				} else {
					pdf.SetFont("Helvetica", "", 12)
				}

				pdf.CellFormat(w, 6, lyricToken.Literal, "0", 0, "L", false, 0, "")
				if len(lyricToken.Chords) > 0 && !p.Config.LyricsOnly {
					postLyricX, postLyricY := pdf.GetXY()
					pdf.SetXY(preX, preY-5)
					for _, chord := range lyricToken.Chords {
						// fmt.Println(chord.Literal)
						cw := pdf.GetStringWidth(chord.Literal)
						pdf.CellFormat(cw, 6, chord.Literal, "0", 0, "", false, 0, "")
						pdf.SetXY(preX+2, preY-5)
					}

					pdf.SetXY(postLyricX, postLyricY)
				}
			}
			// pdf.CellFormat(200, 6, lyrics, "0", 0, "", false, 0, "")
			pdf.Ln(10)
		case types.LineEnvironmentDirective:
			if line.EnvironmentDirective.Name == "soc" {
				chorusEnvironment = true
			}
			if line.EnvironmentDirective.Name == "eoc" {
				chorusEnvironment = false
			}

		default:
		}
	}

	// how the perl project does it
	// https://github.com/ChordPro/chordpro/blob/24ee42236c093063c03cda9f7545e9b9fd117c1a/lib/App/Music/ChordPro/Output/PDF.pm#L1292

	return pdf.Output(w)
}
