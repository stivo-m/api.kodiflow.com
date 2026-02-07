package invoices

import (
	"context"
	"log/slog"

	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/database"
	"github.com/stivo-m/api.kodiflow.com/internal/infrastructure/services/base"
)

type Service interface {
	Generate(ctx context.Context, record *database.Invoice) (pdf []byte, err error)
}

type invoiceService struct {
	log             *slog.Logger
	templateService base.TemplateService
	pdfService      base.PdfService
}

func NewService(templateService base.TemplateService, pdfService base.PdfService) Service {
	logger := slog.Default().WithGroup("metadata").With(slog.String("type", "service"), slog.String("name", "invoices"))
	return invoiceService{
		log:             logger,
		templateService: templateService,
		pdfService:      pdfService,
	}
}

const TemplateFileName = "templates/invoice.html"

func (s invoiceService) Generate(ctx context.Context, record *database.Invoice) (pdf []byte, err error) {

	html, err := s.templateService.Process(ctx, TemplateFileName, record)

	if err != nil {
		s.log.ErrorContext(ctx, "Error while applying template")
		return nil, err
	}

	s.log.InfoContext(ctx, "Template applied")

	pdf, err = s.pdfService.Process(ctx, html)

	if err != nil {
		s.log.ErrorContext(ctx, "Error while converting html template to pdf")
		return nil, err
	}

	return
}

