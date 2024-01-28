package receipt

import (
	"encoding/json"
	"github.com/splitt-org/api/wrappers/http"
	"github.com/splitt-org/api/wrappers/ocr"
	"net/http"
	"regexp"
	"strings"
)

type ErrorDetails struct {
	Message string `json:"message"`
}

type Response struct {
	Success bool          `json:"success"`
	Data    []Item        `json:"data,omitempty"`
	Error   *ErrorDetails `json:"error,omitempty"`
}

type RequestBody struct {
	Image string `json:"image"`
}

type Item struct {
	Name  string `json:"name"`
	Price string `json:"price"`
}

func isPrice(s string) bool {
	regex := regexp.MustCompile(`^\d*\.\d{2}$`)
	return regex.MatchString(s)
}

func findItem(input string) (string, string) {
	parts := strings.Fields(input)
	var name string
	var price string

	for i := len(parts) - 1; i >= 0; i-- {
		if isPrice(parts[i]) {
			price = parts[i]
			name = strings.Join(parts[:i], " ")
			break
		}
	}

	return name, price
}

func mergeOCRlines(ocrLines []splittocr.OCRLine) map[int]string {
	mergedLines := make(map[int]string)

	for _, line := range ocrLines {
		if len(line.Words) == 0 {
			continue
		}

		topValue := int(line.Words[0].Top)
		found := false

		for mergedTop := range mergedLines {
			if topValue >= mergedTop-1 && topValue <= mergedTop+1 {
				mergedLines[mergedTop] += " " + line.LineText
				found = true
				break
			}
		}

		if !found {
			mergedLines[topValue] = line.LineText
		}
	}

	return mergedLines
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
		"base64Image": "data:image/png;base64," + image,
		"isTable":     "true",
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

	var ocrRes splittocr.OCRResponse
	if err := json.Unmarshal(responseData, &ocrRes); err != nil {
		crw.SendJSONResponse(http.StatusInternalServerError, Response{
			Success: false,
			Error: &ErrorDetails{
				Message: "Failed to parse OCR response.",
			},
		})
		return
	}

	lines := ocrRes.ParsedResults[0].TextOverlay.Lines
	linesByTop := mergeOCRlines(lines)

	var items []Item

	for _, line := range linesByTop {
		name, price := findItem(line)
		if name != "" && price != "" {
			items = append(items, Item{Name: name, Price: price})
		}
	}

	crw.SendJSONResponse(http.StatusOK, Response{
		Success: true,
		Data:    items,
	})
}
