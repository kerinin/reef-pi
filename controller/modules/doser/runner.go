package doser

import (
	"fmt"
	"log"
	"time"

	"github.com/kerinin/gomata"
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
		return r.timeDose(cfg, duration)
	}

	if cfg := r.pump.FirmataStepsConfig; cfg != nil {
		return r.stepDose(cfg, int32(volume))
	}

	return fmt.Errorf("ERROR: dosing sub-system. Unconfigured pump")
}

func (r *Runner) timeDose(cfg *TimeConfig, duration float64) error {
	v := make(map[int]float64)
	v[cfg.Pin] = cfg.Speed
	if err := r.jacks.Control(cfg.Jack, v); err != nil {
		return err
	}
	select {
	case <-time.After(time.Duration(duration * float64(time.Second))):
		v[cfg.Pin] = 0
		if err := r.jacks.Control(cfg.Jack, v); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) stepDose(cfg *FirmataStepsConfig, steps int32) error {
	if r.firmata == nil {
		fmt.Errorf("ERROR: dosing sub-system: Firmata not configured for step dosing")
	}
	if err := r.firmata.StepperConfigure(
		cfg.DeviceID,
		gomata.WireCount(cfg.WireCount),
		gomata.StepType(cfg.StepType),
		gomata.HasEnablePin(cfg.HasEnable),
		cfg.Pin1,
		cfg.Pin2,
		cfg.Pin3,
		cfg.Pin4,
		cfg.EnablePin,
		gomata.Inversions(cfg.Invert),
	); err != nil {
		return err
	}
	if err := r.firmata.SetAcceleration(cfg.DeviceID, cfg.Acceleration); err != nil {
		return err
	}
	if err := r.firmata.SetSpeed(cfg.DeviceID, cfg.Speed); err != nil {
		return err
	}
	if err := r.firmata.Step(cfg.DeviceID, steps); err != nil {
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
