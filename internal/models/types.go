package models

type KeyListResponse struct {
	Keys   []string `json:"keys"`
	Total  int      `json:"total"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
}

type ValueResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value"`      // UTF-8 or Base64
	ValueHex string `json:"value_hex"`  // Hex representation
	Size     int    `json:"size"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatsResponse struct {
	TotalKeys   int    `json:"total_keys"`
	DBPath      string `json:"db_path"`
	DBSizeBytes int64  `json:"db_size_bytes"`
}
