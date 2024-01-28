package receipt

import (
	"encoding/json"
	"fmt"
	"github.com/splitt-org/api/wrappers/http"
	"github.com/splitt-org/api/wrappers/ocr"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type ErrorDetails struct {
	Message string `json:"message"`
}

type ReceiptData struct {
	Items []Item `json:"items"`
	Tax   string `json:"tax"`
	Tip   string `json:"tip"`
	Total string `json:"total"`
}

type Response struct {
	Success bool          `json:"success"`
	Data    ReceiptData   `json:"data,omitempty"`
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
	regex := regexp.MustCompile(`^\$?\d*\.\d{2}$`)
	return regex.MatchString(s)
}

func containsAsWord(s string, target string) bool {
	escapedTarget := regexp.QuoteMeta(target)
	regexPattern := fmt.Sprintf(`(?i)\b%s\b`, escapedTarget)
	regex := regexp.MustCompile(regexPattern)

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

type MergedLine struct {
	Top      int
	LineText string
}

func mergeOCRlines(ocrLines []splittocr.OCRLine) []MergedLine {
	var mergedLines []MergedLine

	for _, line := range ocrLines {
		if len(line.Words) == 0 {
			continue
		}

		topValue := int(line.Words[0].Top)
		found := false

		for i := range mergedLines {
			if topValue >= mergedLines[i].Top-1 && topValue <= mergedLines[i].Top+1 {
				mergedLines[i].LineText += " " + line.LineText
				found = true
				break
			}
		}

		if !found {
			mergedLines = append(mergedLines, MergedLine{
				Top:      topValue,
				LineText: line.LineText,
			})
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
		"apikey": "K85545173588957",
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

  log.Println(responseData)

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

	if len(ocrRes.ParsedResults) == 0 {
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
	tax := "0.00"
	tip := "0.00"
	total := "0.00"

	for _, line := range linesByTop {
		name, price := findItem(line.LineText)

		if name == "" || price == "" || containsAsWord(name, "subtotal") {
			continue
		}

		if containsAsWord(name, "tax") {
			tax = price
			continue
		}

		if containsAsWord(name, "tip") {
			tip = price
			continue
		}

		if containsAsWord(name, "total") {
			total = price
			continue
		}

		items = append(items, Item{Name: name, Price: price})
	}

	crw.SendJSONResponse(http.StatusOK, Response{
		Success: true,
		Data: ReceiptData{
			Items: items,
			Tax:   tax,
			Tip:   tip,
			Total: total,
		},
	})
}
