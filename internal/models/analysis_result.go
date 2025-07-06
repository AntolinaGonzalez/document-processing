package models

type FileAnalysisResult struct {
	FileName      string         `json:"file_name"`
	WordCount     int            `json:"word_count"`
	LineCount     int            `json:"line_count"`
	CharCount     int            `json:"char_count"`
	FrequentWords map[string]int `json:"frequent_words"`
	Summary       string         `json:"summary"`
}

type AnalysisResult struct {
	ProcessID string               `json:"process_id"`
	Results   []FileAnalysisResult `json:"results"`
}
