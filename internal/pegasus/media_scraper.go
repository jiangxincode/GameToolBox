package pegasus

// MediaScrapeResult is the result summary for scraping media files.
//
// Downloaded files are counted as Created; existing files skipped are Skipped.
// Any failures are collected in Errors and accounted in Failed.
type MediaScrapeResult struct {
	Created int
	Skipped int
	Failed  int

	Errors []error
}
