package connectors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kerinin/gomata"
	"github.com/reef-pi/reef-pi/controller/storage"
	"github.com/reef-pi/reef-pi/controller/utils"
	"github.com/tarm/serial"
)

type Firmata struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	SerialPath string `json:"serialPath"`
	BaudRate   int    `json:"baudRate"`
	f          *gomata.Firmata
}

func (f *Firmata) IsValid() error {
	if f.SerialPath == "" {
		return fmt.Errorf("Serial path can not be empty")
	}
	if _, err := os.Stat(f.SerialPath); os.IsNotExist(err) {
		return fmt.Errorf("Serial path does not exist")
	}
	if f.BaudRate == 0 {
		return fmt.Errorf("Baud Rate must be nonzero")
	}
	return nil
}

func (f *Firmata) Setup() error {
	port, err := serial.OpenPort(&serial.Config{Name: f.SerialPath, Baud: f.BaudRate})
	if err != nil {
		return err
	}

	f.f = gomata.New()

	err = f.f.Connect(port)
	if err != nil {
		return err
	}

	return nil
}

func (f *Firmata) StepperConfigure(devID int, wireCount gomata.WireCount, stepType gomata.StepType, hasEnable gomata.HasEnablePin, pin1 int, pin2 int, pin3 int, pin4 int, enablePin int, invert gomata.Inversions) error {
	return nil
}

// StepperStep (relative mode)
// Steps to move is specified as a 32-bit signed long.
func (f *Firmata) Step(devID int, v int32) error {
	return nil
}

// StepperSetAcceleration sets the stepper's acceleration
// Sets the acceleration/deceleration in steps/sec^2. The accel value is passed
// using accelStepperFirmata's custom float format
func (f *Firmata) SetAcceleration(devID int, v float32) error {
	return nil
}

// StepperSetSpeed sets the stepper's (max) speed
// If acceleration is off (equal to zero) sets the speed in steps per second.
// If acceleration is on (non-zero) sets the maximum speed in steps per second.
// The speed value is passed using accelStepperFirmata's custom float format.
func (f *Firmata) SetSpeed(devID int, v float32) error {
	return nil
}

type Firmatas struct {
	store storage.Store
}

func NewFirmatas(store storage.Store) *Firmatas {
	return &Firmatas{store}
}

// Connect connects to the Firmata given conn. It first resets the firmata board
// then continuously polls the firmata board for new information when it's
// available.
func (c *Firmatas) Setup() error {
	if err := c.store.CreateBucket(storage.FirmataBucket); err != nil {
		return err
	}

	return nil
}

func (c *Firmatas) Get(id string) (*Firmata, error) {
	var f Firmata
	return &f, c.store.Get(storage.FirmataBucket, id, &f)
}

func (c *Firmatas) List() ([]Firmata, error) {
	firmatas := []Firmata{}
	fn := func(_ string, v []byte) error {
		var f Firmata
		if err := json.Unmarshal(v, &f); err != nil {
			return err
		}
		firmatas = append(firmatas, f)
		return nil
	}
	return firmatas, c.store.List(storage.FirmataBucket, fn)
}

func (c *Firmatas) Create(f Firmata) error {
	if err := f.IsValid(); err != nil {
		return err
	}

	fn := func(id string) interface{} {
		f.ID = id
		return &f
	}
	if err := c.store.Create(storage.FirmataBucket, fn); err != nil {
		return err
	}
	return nil
}

func (c *Firmatas) Update(id string, f Firmata) error {
	if err := f.IsValid(); err != nil {
		return err
	}
	f.ID = id
	if err := c.store.Update(storage.FirmataBucket, id, f); err != nil {
		return err
	}
	return nil
}

func (c *Firmatas) Delete(id string) error {
	_, err := c.Get(id)
	if err != nil {
		return err
	}
	return c.store.Delete(storage.FirmataBucket, id)
}

func (c *Firmatas) LoadAPI(r *mux.Router) {

	// swagger:route GET /api/firmatas Firmata firmataList
	// List all firmatas.
	// List all firmatas in reef-pi.
	// responses:
	// 	200: body:[]firmata
	// 	500:
	r.HandleFunc("/api/firmatas", c.list).Methods("GET")

	// swagger:operation GET /api/firmatas/{id} Firmata firmataGet
	// Get a Firmata by id.
	// Get an existing Firmata.
	// ---
	// parameters:
	//  - in: path
	//    name: id
	//    description: The Id of the firmata
	//    required: true
	//    schema:
	//     type: integer
	// responses:
	//  200:
	//   description: OK
	//   schema:
	//    $ref: '#/definitions/firmata'
	//  404:
	//   description: Not Found
	r.HandleFunc("/api/firmatas/{id}", c.get).Methods("GET")

	// swagger:operation PUT /api/firmatas Firmata firmataCreate
	// Create a Firmata.
	// Create a new Firmata.
	// ---
	// parameters:
	//  - in: body
	//    name: firmata
	//    description: The firmata to create
	//    required: true
	//    schema:
	//     $ref: '#/definitions/firmata'
	// responses:
	//  200:
	//   description: OK
	r.HandleFunc("/api/firmatas", c.create).Methods("PUT")

	//swagger:operation POST /api/firmatas/{id} Firmata firmataUpdate
	// Update a Firmata.
	// Update an existing Firmata.
	//
	//---
	//parameters:
	// - in: path
	//   name: id
	//   description: The Id of the firmata to update
	//   required: true
	//   schema:
	//    type: integer
	// - in: body
	//   name: firmata
	//   description: The firmata to update
	//   required: true
	//   schema:
	//    $ref: '#/definitions/firmata'
	//responses:
	// 200:
	//  description: OK
	// 404:
	//  description: Not Found
	r.HandleFunc("/api/firmatas/{id}", c.update).Methods("POST")

	// swagger:operation DELETE /api/firmatas/{id} Firmata firmataDelete
	// Delete a Firmata.
	// Delete an existing Firmata.
	// ---
	// parameters:
	//  - in: path
	//    name: id
	//    description: The Id of the firmata to delete
	//    required: true
	//    schema:
	//     type: integer
	// responses:
	//  200:
	//   description: OK
	r.HandleFunc("/api/firmatas/{id}", c.delete).Methods("DELETE")
}

func (c *Firmatas) get(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) (interface{}, error) {
		return c.Get(id)
	}
	utils.JSONGetResponse(fn, w, r)
}

func (c *Firmatas) list(w http.ResponseWriter, r *http.Request) {
	fn := func() (interface{}, error) {
		return c.List()
	}
	utils.JSONListResponse(fn, w, r)
}

func (c *Firmatas) create(w http.ResponseWriter, r *http.Request) {
	var f Firmata
	fn := func() error {
		return c.Create(f)
	}
	utils.JSONCreateResponse(&f, fn, w, r)
}

func (c *Firmatas) update(w http.ResponseWriter, r *http.Request) {
	var f Firmata
	fn := func(id string) error {
		return c.Update(id, f)
	}
	utils.JSONUpdateResponse(&f, fn, w, r)
}

func (c *Firmatas) delete(w http.ResponseWriter, r *http.Request) {
	fn := func(id string) error {
		return c.Delete(id)
	}
	utils.JSONDeleteResponse(fn, w, r)
}
