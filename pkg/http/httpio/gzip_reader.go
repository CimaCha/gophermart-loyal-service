// Package httpio предоставляет инструменты для работы с вводом-выводом в HTTP-контексте,
// включая декомпрессию данных на лету.
package httpio

import (
	"compress/gzip"
	"errors"
	"io"
)

type gzipReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewGzipReader создает и инициализирует новый gzipReader, который автоматически
// распаковывает поток данных из переданного io.ReadCloser.
// Возвращает ошибку, если gzip-заголовок поврежден или невалиден.
func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read считывает распакованные байты в переданный срез p.
// Соответствует интерфейсу io.Reader.
func (gr *gzipReader) Read(p []byte) (n int, err error) {
	return gr.zr.Read(p)
}

// Close закрывает как исходный поток данных, так и внутренний gzip-читатель.
// Объединяет ошибки закрытия, если они возникают. Соответствует интерфейсу io.Closer.
func (gr *gzipReader) Close() error {
	var errs error
	if err := gr.r.Close(); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := gr.zr.Close(); err != nil {
		errs = errors.Join(errs, err)
	}
	if errs != nil {
		return errs
	}
	return nil
}
