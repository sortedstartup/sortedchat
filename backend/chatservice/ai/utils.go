package ai

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ImageFormat represents the format of an image
type ImageFormat string

const (
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatWEBP ImageFormat = "webp"
	ImageFormatGIF  ImageFormat = "gif"
)

// ImageData represents image data with metadata
type ImageData struct {
	Data     []byte      `json:"data"`
	Format   ImageFormat `json:"format"`
	Base64   string      `json:"base64"`
	MimeType string      `json:"mime_type"`
}

// ParseMultimodalInput parses text input that may contain image references
// Supports formats like:
// - [image:base64:data] - inline base64 image data
// - [image:url:https://...] - image URL
// - [image:file:path] - local file path (for server-side processing)
func ParseMultimodalInput(text string) ([]MessageContent, error) {
	var contents []MessageContent

	// Regex to match image references
	imageRegex := regexp.MustCompile(`\[image:(base64|url|file):([^\]]+)\]`)

	lastIndex := 0
	matches := imageRegex.FindAllStringSubmatchIndex(text, -1)

	for _, match := range matches {
		// Add text content before the image
		if match[0] > lastIndex {
			textContent := strings.TrimSpace(text[lastIndex:match[0]])
			if textContent != "" {
				contents = append(contents, NewTextContent(textContent))
			}
		}

		// Extract image type and data
		imageType := text[match[2]:match[3]]
		imageData := text[match[4]:match[5]]

		switch imageType {
		case "base64":
			contents = append(contents, NewImageContent("data:image/jpeg;base64,"+imageData))
		case "url":
			// For URLs, we might want to download and encode, or pass through
			contents = append(contents, NewImageContent(imageData))
		case "file":
			// For file paths, we'd need to read and encode the file
			imageContent, err := ProcessImageFile(imageData)
			if err != nil {
				return nil, fmt.Errorf("failed to process image file %s: %v", imageData, err)
			}
			contents = append(contents, imageContent)
		}

		lastIndex = match[1]
	}

	// Add remaining text after the last image
	if lastIndex < len(text) {
		textContent := strings.TrimSpace(text[lastIndex:])
		if textContent != "" {
			contents = append(contents, NewTextContent(textContent))
		}
	}

	// If no images found, just return the text as-is
	if len(contents) == 0 {
		contents = append(contents, NewTextContent(text))
	}

	return contents, nil
}

// ProcessImageFile reads an image file and converts it to base64
func ProcessImageFile(filePath string) (MessageContent, error) {
	// This would be implemented to read local files
	// For security reasons, you might want to restrict file paths
	return MessageContent{}, fmt.Errorf("file processing not implemented - use base64 or URL instead")
}

// ProcessImageURL downloads an image from URL and converts to base64
func ProcessImageURL(url string) (*ImageData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v", err)
	}

	// Detect image format from content type
	contentType := resp.Header.Get("Content-Type")
	format, mimeType := detectImageFormat(contentType, data)

	base64Data := base64.StdEncoding.EncodeToString(data)

	return &ImageData{
		Data:     data,
		Format:   format,
		Base64:   base64Data,
		MimeType: mimeType,
	}, nil
}

// detectImageFormat detects image format from content type or file data
func detectImageFormat(contentType string, data []byte) (ImageFormat, string) {
	// Check content type first
	switch contentType {
	case "image/jpeg":
		return ImageFormatJPEG, "image/jpeg"
	case "image/png":
		return ImageFormatPNG, "image/png"
	case "image/webp":
		return ImageFormatWEBP, "image/webp"
	case "image/gif":
		return ImageFormatGIF, "image/gif"
	}

	// Fallback to file signature detection
	if len(data) >= 4 {
		// JPEG signature
		if data[0] == 0xFF && data[1] == 0xD8 {
			return ImageFormatJPEG, "image/jpeg"
		}
		// PNG signature
		if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
			return ImageFormatPNG, "image/png"
		}
		// WebP signature
		if len(data) >= 12 && string(data[8:12]) == "WEBP" {
			return ImageFormatWEBP, "image/webp"
		}
		// GIF signature
		if string(data[0:3]) == "GIF" {
			return ImageFormatGIF, "image/gif"
		}
	}

	// Default to JPEG
	return ImageFormatJPEG, "image/jpeg"
}

// ConvertToBase64DataURL converts image data to a data URL format
func ConvertToBase64DataURL(imageData *ImageData) string {
	return fmt.Sprintf("data:%s;base64,%s", imageData.MimeType, imageData.Base64)
}

// ExtractBase64FromDataURL extracts base64 data from a data URL
func ExtractBase64FromDataURL(dataURL string) (string, string, error) {
	if !strings.HasPrefix(dataURL, "data:") {
		return "", "", fmt.Errorf("invalid data URL format")
	}

	parts := strings.SplitN(dataURL, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid data URL format")
	}

	header := parts[0]
	base64Data := parts[1]

	// Extract mime type
	mimeType := ""
	if strings.Contains(header, ";") {
		mimeType = strings.TrimPrefix(strings.Split(header, ";")[0], "data:")
	}

	return base64Data, mimeType, nil
}

// ValidateImageSize checks if image size is within limits
func ValidateImageSize(data []byte, maxSizeMB int) error {
	sizeMB := len(data) / (1024 * 1024)
	if sizeMB > maxSizeMB {
		return fmt.Errorf("image size %d MB exceeds limit of %d MB", sizeMB, maxSizeMB)
	}
	return nil
}

// CreateMultimodalMessage creates a multimodal message with both text and images
func CreateMultimodalMessage(role, text string, images []string) ChatMessage {
	contents := []MessageContent{}

	if text != "" {
		contents = append(contents, NewTextContent(text))
	}

	for _, imageURL := range images {
		contents = append(contents, NewImageContent(imageURL))
	}

	return NewMultimodalMessage(role, contents)
}

// ProviderSpecificFormatting contains provider-specific formatting logic
type ProviderSpecificFormatting struct{}

// FormatForOpenAI formats content for OpenAI's Responses API
func (f *ProviderSpecificFormatting) FormatForOpenAI(content MessageContent) map[string]interface{} {
	switch content.Type {
	case ContentTypeText:
		return map[string]interface{}{
			"type": "input_text",
			"text": content.Text,
		}
	case ContentTypeImage:
		return map[string]interface{}{
			"type":      "input_image",
			"image_url": content.ImageURL,
		}
	}
	return nil
}

// FormatForClaude formats content for Claude's API
func (f *ProviderSpecificFormatting) FormatForClaude(content MessageContent) map[string]interface{} {
	switch content.Type {
	case ContentTypeText:
		return map[string]interface{}{
			"type": "text",
			"text": content.Text,
		}
	case ContentTypeImage:
		// Claude expects base64 data without the data URL prefix
		base64Data, mimeType, err := ExtractBase64FromDataURL(content.ImageURL)
		if err != nil {
			// If not a data URL, assume it's already base64
			base64Data = content.ImageURL
			mimeType = "image/jpeg"
		}

		return map[string]interface{}{
			"type": "image",
			"source": map[string]interface{}{
				"type":       "base64",
				"media_type": mimeType,
				"data":       base64Data,
			},
		}
	}
	return nil
}

// FormatForGemini formats content for Gemini's API
func (f *ProviderSpecificFormatting) FormatForGemini(content MessageContent) map[string]interface{} {
	switch content.Type {
	case ContentTypeText:
		return map[string]interface{}{
			"text": content.Text,
		}
	case ContentTypeImage:
		// Gemini expects base64 data without the data URL prefix
		base64Data, mimeType, err := ExtractBase64FromDataURL(content.ImageURL)
		if err != nil {
			// If not a data URL, assume it's already base64
			base64Data = content.ImageURL
			mimeType = "image/jpeg"
		}

		return map[string]interface{}{
			"inline_data": map[string]interface{}{
				"mime_type": mimeType,
				"data":      base64Data,
			},
		}
	}
	return nil
}
