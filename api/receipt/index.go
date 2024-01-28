package receipt

import (
	"github.com/splitt-org/api/wrappers/http"
	"github.com/splitt-org/api/wrappers/ocr"
	"log"
	"net/http"
)

type ErrorDetails struct {
	Message string `json:"message"`
}

type Response struct {
	Success bool          `json:"success"`
	Data    string        `json:"data,omitempty"`
	Error   *ErrorDetails `json:"error,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	crw := &splitthttp.ResponseWriter{W: w}
	crw.SetCors(r.Host)

	headers := map[string]string{
		"apikey": "helloworld",
	}

	formValues := map[string]string{
		"url":     "https://ocr.space/Content/Images/receipt-ocr-original.jpg",
		"isTable": "true",
	}

	ocrReq := splittocr.NewOCRRequest(headers, formValues)

	responseData, err := splittocr.PostOCRRequest(ocrReq)
	if err != nil {
		crw.SendJSONResponse(http.StatusOK, Response{
			Success: false,
			Error: &ErrorDetails{
				Message: "Failed to OCR.",
			},
		})
		return
	}

	log.Println(string(responseData))
	crw.SendJSONResponse(http.StatusOK, Response{
		Success: true,
		Data:    string(responseData),
	})
}
