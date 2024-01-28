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

type OCRWord struct {
    WordText string  `json:"WordText"`
    Left     float64 `json:"Left"`
    Top      float64 `json:"Top"`
    Height   float64 `json:"Height"`
    Width    float64 `json:"Width"`
}

type OCRLine struct {
    LineText  string     `json:"LineText"`
    Words     []OCRWord  `json:"Words"`
    MaxHeight float64    `json:"MaxHeight"`
    MinTop    float64    `json:"MinTop"`
}

type TextOverlay struct {
    Lines       []OCRLine `json:"Lines"`
    HasOverlay  bool      `json:"HasOverlay"`
    Message     string    `json:"Message"`
}

type ParsedResult struct {
    TextOverlay           TextOverlay `json:"TextOverlay"`
    TextOrientation       string      `json:"TextOrientation"`
    FileParseExitCode     int         `json:"FileParseExitCode"`
    ParsedText            string      `json:"ParsedText"`
    ErrorMessage          string      `json:"ErrorMessage"`
    ErrorDetails          string      `json:"ErrorDetails"`
}

type OCRResponse struct {
    ParsedResults                 []ParsedResult `json:"ParsedResults"`
    OCRExitCode                   int            `json:"OCRExitCode"`
    IsErroredOnProcessing         bool           `json:"IsErroredOnProcessing"`
    ProcessingTimeInMilliseconds  string         `json:"ProcessingTimeInMilliseconds"`
    SearchablePDFURL              string         `json:"SearchablePDFURL"`
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
