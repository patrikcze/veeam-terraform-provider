package models

// SecurityAnalyzerScheduleModel is the API request/response body for the security analyzer schedule singleton.
type SecurityAnalyzerScheduleModel struct {
	RunAutomatically bool                             `json:"runAutomatically"`
	Daily            *SecurityAnalyzerDailySchedule   `json:"daily,omitempty"`
	Monthly          *SecurityAnalyzerMonthlySchedule `json:"monthly,omitempty"`
}

// SecurityAnalyzerDailySchedule holds the daily schedule configuration.
type SecurityAnalyzerDailySchedule struct {
	IsEnabled bool   `json:"isEnabled"`
	LocalTime string `json:"localTime,omitempty"`
}

// SecurityAnalyzerMonthlySchedule holds the monthly schedule configuration.
type SecurityAnalyzerMonthlySchedule struct {
	IsEnabled  bool `json:"isEnabled"`
	DayOfMonth int  `json:"dayOfMonth,omitempty"`
}
