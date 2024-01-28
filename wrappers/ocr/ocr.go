package splittocr

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	URL = "https://api.ocr.space/parse/image"
)

type OCRRequest struct {
	Headers    map[string]string
	FormValues map[string]string
}

func NewOCRRequest(headers map[string]string, formValues map[string]string) *OCRRequest {
	return &OCRRequest{
		Headers:    headers,
		FormValues: formValues,
	}
}

func PostOCRRequest(ocrReq *OCRRequest) ([]byte, error) {
	formData := url.Values{}
	for key, value := range ocrReq.FormValues {
		formData.Set(key, value)
	}

	request, err := http.NewRequest("POST", URL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	for headerKey, headerValue := range ocrReq.Headers {
		request.Header.Set(headerKey, headerValue)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	return io.ReadAll(response.Body)
}
