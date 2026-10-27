package app

// reviewConfig is what the scan and the review of a request need to know
// about the working copy.
type reviewConfig struct {
	header, idHeader string
	overviewDir      string
	sheetRel         string            // rel path of the overview spreadsheet
	template         string            // rel path of the template
	templateBase     string            // its name without extension
	known            map[string]bool   // checksums of the forms numb-form wrote
	names            map[string]string // folder name (lower case) -> company, from the spreadsheet
	ids              map[string]string // folder name (lower case) -> customer id
}
