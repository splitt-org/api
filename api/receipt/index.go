package receipt

import (
	"github.com/splitt-org/api/wrappers/http"
	"github.com/splitt-org/api/wrappers/ocr"
	"net/http"
  "json"
)

type ErrorDetails struct {
	Message string `json:"message"`
}

type Response struct {
	Success bool          `json:"success"`
	Data    []byte        `json:"data,omitempty"`
	Error   *ErrorDetails `json:"error,omitempty"`
}

type RequestBody struct {
    Image string `json:"image"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	crw := &splitthttp.ResponseWriter{W: w}
	crw.SetCors(r.Host)

  var reqBody RequestBody
  if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
      crw.SendJSONResponse(http.StatusBadRequest, Response{
        Success: false,
        Error: &ErrorDetails{
          Message: "Invalid request body.",
        },
      })
      return
  }

  image := reqBody.Image

  if image == "" {
    crw.SendJSONResponse(http.StatusBadRequest, Response{
			Success: false,
			Error: &ErrorDetails{
				Message: "No image query is populated.",
			},
		})
		return
  }

	headers := map[string]string{
		"apikey": "helloworld",
	}

	formValues := map[string]string{
		"base64Image": image,
		"isTable": "true",
	}

	ocrReq := splittocr.NewOCRRequest(headers, formValues)

	responseData, err := splittocr.PostOCRRequest(ocrReq)
	if err != nil {
		crw.SendJSONResponse(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorDetails{
				Message: "Failed to OCR.",
			},
		})
		return
	}

	crw.SendJSONResponse(http.StatusOK, Response{
		Success: true,
		Data:    responseData,
	})
}
