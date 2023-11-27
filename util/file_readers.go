package util

import (
	"bytes"
	"github.com/klippa-app/go-pdfium"
	"github.com/klippa-app/go-pdfium/requests"
	"github.com/klippa-app/go-pdfium/webassembly"
	"os"
	"sync"
	"time"
)

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

func ReadPDF(pdfFilePath string) (string, error) {
	initPDFInstance()
	// Load the PDF file into a byte array.
	pdfBytes, err := os.ReadFile(pdfFilePath)
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
