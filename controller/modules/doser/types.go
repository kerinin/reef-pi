package doser

import (
	"strings"

	cron "github.com/robfig/cron/v3"
)

//swagger:model dosingRegiment
type DosingRegiment struct {
	Enable   bool     `json:"enable"`
	Schedule Schedule `json:"schedule"`
	// The volume to be dosed. Value to be configured through calibration.
	Volume float64 `json:"volume"`
}

//swagger:model doserCalibrationDetails
type CalibrationDetails struct {
	Volume float64 `json:"volume"`
}

type CalibrationResult struct {
	Details        CalibrationDetails `json:"details"`
	MeasuredVolume float64            `json:"measuredVolume"`
}

type Schedule struct {
	Day    string `json:"day"`
	Hour   string `json:"hour"`
	Minute string `json:"minute"`
	Second string `json:"second"`
	Week   string `json:"week"`
	Month  string `json:"month"`
}

func (s Schedule) CronSpec() string {
	return strings.Join([]string{s.Second, s.Minute, s.Hour, s.Day, s.Month, s.Week}, " ")
}

func (s Schedule) Validate() error {
	parser := cron.NewParser(_cronParserSpec)
	_, err := parser.Parse(s.CronSpec())
	return err
}
