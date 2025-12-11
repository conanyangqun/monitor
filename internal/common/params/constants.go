package params

// Character sets for random string generation
const (
	DefaultCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	HexCharset     = "0123456789abcdef"
)

// TimeFormatReplacements Time format pattern replacements for strftime-like patterns to Go time format
var TimeFormatReplacements = map[string]string{
	"%Y": "2006",    // 4-digit year
	"%y": "06",      // 2-digit year
	"%m": "01",      // month (01-12)
	"%d": "02",      // day (01-31)
	"%H": "15",      // hour (00-23)
	"%M": "04",      // minute (00-59)
	"%S": "05",      // second (00-59)
	"%B": "January", // full month name
	"%b": "Jan",     // abbreviated month name
	"%A": "Monday",  // full weekday name
	"%a": "Mon",     // abbreviated weekday name
	"%j": "002",     // day of year (001-366)
	"%U": "",        // week of year (placeholder)
	"%W": "",        // week of year (placeholder)
	"%w": "",        // weekday (placeholder)
	"%Z": "MST",     // timezone name
	"%z": "-0700",   // timezone offset
	"%s": "",        // Unix timestamp (placeholder)
}

// HTTPMethods HTTP methods for random HTTP method generation
var HTTPMethods = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
	"HEAD",
	"OPTIONS",
}

// UserAgents User agent strings for random user agent generation
var UserAgents = LoadUserAgents()

// MimeTypes MIME types for random MIME type generation
var MimeTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/webp",
	"video/mp4",
	"video/x-msvideo",
	"video/x-flv",
	"audio/mpeg",
	"audio/ogg",
	"application/pdf",
	"application/zip",
	"application/x-rar-compressed",
	"text/html",
	"text/css",
	"text/javascript",
	"application/json",
	"application/xml",
}

// FileExtensions File extensions for random file extension generation
var FileExtensions = []string{
	".jpg",
	".jpeg",
	".png",
	".gif",
	".webp",
	".mp4",
	".avi",
	".flv",
	".mp3",
	".ogg",
	".pdf",
	".zip",
	".rar",
	".html",
	".css",
	".js",
	".json",
	".xml",
}

// FirstNames First names for fake name generation
var FirstNames = LoadFirstNames()

// LastNames Last names for fake name generation
var LastNames = LoadLastNames()

// FakeDomains Domain names for fake domain generation
var FakeDomains = LoadFakeDomains()
