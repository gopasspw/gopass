package cairo

// MimeType is a representation of Cairo's CAIRO_MIME_TYPE_*
// preprocessor constants.
type MimeType string

const (
	MIME_TYPE_JP2       MimeType = "image/jp2"
	MIME_TYPE_JPEG      MimeType = "image/jpeg"
	MIME_TYPE_PNG       MimeType = "image/png"
	MIME_TYPE_URI       MimeType = "image/x-uri"
	MIME_TYPE_UNIQUE_ID MimeType = "application/x-cairo.uuid"
)
