package model

type LogEvent struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

func (e *LogEvent) GetKey() string {
	return e.Service
}

func (e *LogEvent) GetId() int {
	return 0
}
