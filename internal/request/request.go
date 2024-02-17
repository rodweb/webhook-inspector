package request

type Request struct {
	ID        string              `json:"id"`
	Method    string              `json:"method"`
	Endpoint  string              `json:"endpoint"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	Timestamp int64               `json:"timestamp"`
}
