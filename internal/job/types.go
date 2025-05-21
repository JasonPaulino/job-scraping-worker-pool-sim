package job

type Job struct {
	ID  string
	URL string
}

type Result struct {
	JobID      string
	StatusCode int
	Errors     []error
}
