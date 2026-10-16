package model

type CpuRequest struct {
	Iterations int `json:"iterations"`
	HashRounds int `json:"hashRounds"`
	SortSize   int `json:"sortSize"`
}

type CpuResult struct {
	HashAccumulator int64   `json:"hashAccumulator"`
	Median          int     `json:"median"`
	Sum             int64   `json:"sum"`
	ElapsedNs       int64   `json:"elapsedNs"`
	ElapsedMs       float64 `json:"elapsedMs"`
}

type ConcurrencyResult struct {
	Accumulator int64 `json:"accumulator"`
	ElapsedNs   int64 `json:"elapsedNs"`
}

type AllocationRequest struct {
	ObjectCount     int `json:"objectCount"`
	ObjectSizeBytes int `json:"objectSizeBytes"`
	TransformRounds int `json:"transformRounds"`
}

type AllocationResult struct {
	ObjectCount int     `json:"objectCount"`
	TotalBytes  int64   `json:"totalBytes"`
	TotalMB     float64 `json:"totalMB"`
	Checksum    int     `json:"checksum"`
	ElapsedNs   int64   `json:"elapsedNs"`
	ElapsedMs   float64 `json:"elapsedMs"`
}
