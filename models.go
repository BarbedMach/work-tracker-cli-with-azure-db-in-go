package main

type WorkItem struct {
	ID          int    `json:"id"`
	WorkDate    string `json:"workDate"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	Description string `json:"description"`
}
