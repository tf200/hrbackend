package pdf

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"strings"
	"time"

	pkgbucket "hrbackend/pkg/bucket"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

type documentSection struct {
	Title string
	Lines []string
}

func buildSectionsPDF(
	title string,
	headerLines []string,
	sections []documentSection,
) ([]byte, error) {
	cfg := config.NewBuilder().
		WithLeftMargin(12).
		WithRightMargin(12).
		WithTopMargin(14).
		WithBottomMargin(12).
		Build()

	m := maroto.New(cfg)

	m.AddAutoRow(text.NewCol(12, strings.TrimSpace(title), props.Text{
		Style: fontstyle.Bold,
		Size:  16,
		Align: align.Center,
	}))

	m.AddAutoRow(
		text.NewCol(
			12,
			fmt.Sprintf("Generated at: %s", time.Now().Format("2006-01-02 15:04")),
			props.Text{
				Size:  9,
				Align: align.Right,
			},
		),
	)

	for _, line := range headerLines {
		clean := strings.TrimSpace(line)
		if clean == "" {
			continue
		}
		m.AddAutoRow(text.NewCol(12, clean, props.Text{Size: 10}))
	}

	if len(headerLines) > 0 {
		m.AddAutoRow(text.NewCol(12, "", props.Text{Size: 4}))
	}

	for _, section := range sections {
		sectionTitle := strings.TrimSpace(section.Title)
		if sectionTitle != "" {
			m.AddAutoRow(text.NewCol(12, sectionTitle, props.Text{
				Style: fontstyle.Bold,
				Size:  12,
				Top:   1,
			}))
		}

		lineCount := 0
		for _, line := range section.Lines {
			clean := strings.TrimSpace(line)
			if clean == "" {
				continue
			}
			m.AddAutoRow(text.NewCol(12, "- "+clean, props.Text{Size: 10}))
			lineCount++
		}

		if lineCount == 0 {
			m.AddAutoRow(text.NewCol(12, "- N/A", props.Text{Size: 10}))
		}

		m.AddAutoRow(text.NewCol(12, "", props.Text{Size: 4}))
	}

	document, err := m.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate maroto document: %w", err)
	}

	return document.GetBytes(), nil
}

func toMultipartFile(pdfBytes []byte) multipart.File {
	return &pkgbucket.InMemoryFile{Reader: bytes.NewReader(pdfBytes)}
}
