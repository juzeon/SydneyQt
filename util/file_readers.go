package util

import (
	"archive/zip"
	"bytes"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"github.com/microcosm-cc/bluemonday"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type DocumentReader interface {
	Read(filePath string) (string, error)
}

type PDFDocumentReader struct {
}

var pdfInstance pdfium.Pdfium
var initPDFInstance = sync.OnceFunc(func() {
	pool, err := webassembly.Init(webassembly.Config{
		MinIdle:  1,
		MaxIdle:  1,
		MaxTotal: 1,
	})
	if err != nil {
		panic(err)
	}
	instance, err := pool.GetInstance(time.Second * 30)
	if err != nil {
		panic(err)
	}
	pdfInstance = instance
})

func (P PDFDocumentReader) Read(filePath string) (string, error) {
	initPDFInstance()
	// Load the PDF file into a byte array.
	pdfBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Open the PDF using PDFium (and claim a worker)
	doc, err := pdfInstance.OpenDocument(&requests.OpenDocument{
		File: &pdfBytes,
	})
	if err != nil {
		return "", err
	}
	// Always close the document, this will release its resources.
	defer pdfInstance.FPDF_CloseDocument(&requests.FPDF_CloseDocument{
		Document: doc.Document,
	})
	pageCount, err := pdfInstance.FPDF_GetPageCount(&requests.FPDF_GetPageCount{
		Document: doc.Document,
	})
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	for i := 0; i < pageCount.PageCount; i++ {
		pageText, err := pdfInstance.GetPageText(&requests.GetPageText{Page: requests.Page{
			ByIndex: &requests.PageByIndex{
				Document: doc.Document,
				Index:    i,
			},
		}})
		if err != nil {
			return "", err
		}
		_, err = buf.WriteString(pageText.Text + "\n")
		if err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

type DocxDocumentReader struct {
}

func (d DocxDocumentReader) Read(filePath string) (string, error) {
	reader, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer reader.Close()
	docFile, err := reader.Open("word/document.xml")
	if err != nil {
		return "", err
	}
	defer docFile.Close()
	v, err := io.ReadAll(docFile)
	if err != nil {
		return "", err
	}
	text := string(v)
	text = strings.ReplaceAll(text, "<w:p>", "\n<w:p>")
	return bluemonday.StripTagsPolicy().Sanitize(text), nil
}
