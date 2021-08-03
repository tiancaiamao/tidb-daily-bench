package benchdaily

type BenchOutput struct {
	Date   string
	Commit string
	Result []BenchResult
}

type BenchResult struct {
	Name        string
	NsPerOp     int64
	AllocsPerOp int64
	BytesPerOp  int64
}
