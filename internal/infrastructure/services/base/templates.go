package base

import (
	"bytes"
	"context"
	"html/template"
	"log/slog"
)

type TemplateService interface {
	Process(ctx context.Context, templateFileName string, data any) (html string, err error)
}

type templateService struct {
	log *slog.Logger
}

func NewTemplateService() TemplateService {
	logger := slog.Default().WithGroup("metadata").With(slog.String("type", "service"), slog.String("name", "templates"))

	return templateService{log: logger}
}

func (s templateService) Process(ctx context.Context, templateFileName string, data any) (html string, err error) {

	s.log.InfoContext(ctx, "Parsing html template")

	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		s.log.ErrorContext(ctx, "Error while parsing template")
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		s.log.ErrorContext(ctx, "Error while applying data to template")
		return "", err

	}

	return buf.String(), nil
}
