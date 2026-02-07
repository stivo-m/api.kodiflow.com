package base

import (
	"context"
	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"log/slog"
	"strings"
)

type PdfService interface {
	Process(ctx context.Context, html string) (pdf []byte, err error)
}

type pdfService struct {
	log *slog.Logger
}

func NewPdfService() PdfService {
	logger := slog.Default().WithGroup("metadata").With(slog.String("type", "service"), slog.String("name", "pdf"))
	return pdfService{log: logger}
}

func (s pdfService) Process(ctx context.Context, html string) (pdf []byte, err error) {

	page := wkhtmltopdf.NewPageReader(strings.NewReader(html))
	pdfGenerator, err := wkhtmltopdf.NewPDFGenerator()

	if err != nil {
		s.log.ErrorContext(ctx, "Error while creating new pdf generator")
		return nil, err
	}

	pdfGenerator.AddPage(page)

	err = pdfGenerator.Create()
	if err != nil {
		s.log.ErrorContext(ctx, "Error while creating new pdf page")
		return nil, err
	}

	return pdfGenerator.Bytes(), nil
}
