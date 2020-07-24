package doser

import (
	"fmt"
	"log"
	"time"

	"github.com/reef-pi/reef-pi/controller/connectors"
	"github.com/reef-pi/reef-pi/controller/telemetry"
)

type Runner struct {
	pump     *Pump
	jacks    *connectors.Jacks
	firmata  *connectors.Firmata
	statsMgr telemetry.StatsManager
}

func (r *Runner) Dose(volume float64) error {
	if cfg := r.pump.TimeConfig; cfg != nil {
		var duration = volume / cfg.Speed
		return r.timeDose(cfg.Jack, cfg.Pin, cfg.Speed, duration)
	}

	if cfg := r.pump.FirmataStepsConfig; cfg != nil {
		return r.stepDose(cfg.DeviceID, cfg.Acceleration, cfg.Speed, int32(volume))
	}

	return fmt.Errorf("ERROR: dosing sub-system. Unconfigured pump")
}

func (r *Runner) timeDose(jack string, pin int, speed float64, duration float64) error {
	v := make(map[int]float64)
	v[pin] = speed
	if err := r.jacks.Control(jack, v); err != nil {
		return err
	}
	select {
	case <-time.After(time.Duration(duration * float64(time.Second))):
		v[pin] = 0
		if err := r.jacks.Control(jack, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) stepDose(deviceID int, acceleration, speed float32, steps int32) error {
	if r.firmata == nil {
		fmt.Errorf("ERROR: dosing sub-system: Firmata not configured for step dosing")
	}
	if err := r.firmata.SetAcceleration(deviceID, acceleration); err != nil {
		return err
	}
	if err := r.firmata.SetSpeed(deviceID, speed); err != nil {
		return err
	}
	if err := r.firmata.Step(deviceID, steps); err != nil {
		return err
	}

	return nil
}

func (r *Runner) Run() {
	log.Println("doser sub system: scheduled run ", r.pump.Name)

	var volume = r.pump.Regiment.Volume
	if cal := r.pump.Calibration; cal != nil {
		var doseAdjustment = cal.Details.Volume / cal.MeasuredVolume
		volume = r.pump.Regiment.Volume * doseAdjustment
	}

	if err := r.Dose(volume); err != nil {
		log.Println("ERROR: dosing sub-system. Failed to control pump. Error:", err)
		return
	}

	usage := Usage{
		Time: telemetry.TeleTime(time.Now()),
		Pump: int(r.pump.Regiment.Volume),
	}
	r.statsMgr.Update(r.pump.ID, usage)
	r.statsMgr.Save(r.pump.ID)
	//r.Telemetry().EmitMetric("doser"+r.pump.Name+"-usage", usage.Pump)
}
