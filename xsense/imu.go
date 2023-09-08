package xsense

import (
	"context"
	"io"
	"sync"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	slib "github.com/jacobsa/go-serial/serial"
	geo "github.com/kellydunn/golang-geo"
	"github.com/pkg/errors"
	"go.viam.com/utils"

	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
	rutils "go.viam.com/rdk/utils"
)

var Model = resource.NewModel("viam", "sensor", "mti-xsense-200")
var baudRateList = []uint{115200}

func init() {
	resource.RegisterComponent(movementsensor.API, Model,
		resource.Registration[movementsensor.MovementSensor, *Config]{
			Constructor: newXsense,
		})
}

type Config struct {
	SerialPath     string `json:"serial_path"`
	SerialBaudRate int    `json:"serial_baud_rate,omitempty"`
}

// Validate ensures all parts of the config are valid.
func (cfg *Config) Validate(path string) ([]string, error) {
	var deps []string
	if cfg.SerialPath == "" {
		return nil, utils.NewConfigValidationFieldRequiredError(path, "serial_path")
	}
	// Validating baud rate
	if !rutils.ValidateBaudRate(baudRateList, int(cfg.SerialBaudRate)) {
		return nil, utils.NewConfigValidationError(path, errors.Errorf("Baud rate is not in %v", baudRateList))
	}

	return deps, nil
}

type xsense struct {
	resource.Named
	resource.AlwaysRebuild
	magnetometer            r3.Vector
	compassheading          float64
	numBadReadings          uint32
	err                     movementsensor.LastError
	mu                      sync.Mutex
	port                    io.ReadWriteCloser
	cancelFunc              func()
	activeBackgroundWorkers sync.WaitGroup
	logger                  golog.Logger
}

// Close
func (mti *xsense) Close(ctx context.Context) error {
	return nil
}

// CompassHeading
func (mti *xsense) CompassHeading(ctx context.Context, extra map[string]interface{}) (float64, error) {
	return 0, nil
}

// Accuracy unimplemented
func (mti *xsense) Accuracy(ctx context.Context, extra map[string]interface{}) (map[string]float32, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return nil, nil
}

// AngularVelocity unimplemented
func (mti *xsense) AngularVelocity(ctx context.Context, extra map[string]interface{}) (spatialmath.AngularVelocity, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return spatialmath.AngularVelocity{}, nil
}

// LinearAcceleration unimplemented
func (mti *xsense) LinearAcceleration(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return r3.Vector{}, nil
}

// LinearVelocity unimplemented
func (mti *xsense) LinearVelocity(ctx context.Context, extra map[string]interface{}) (r3.Vector, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return r3.Vector{}, nil
}

// Orientation unimplemented
func (mti *xsense) Orientation(ctx context.Context, extra map[string]interface{}) (spatialmath.Orientation, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return spatialmath.NewZeroOrientation(), nil
}

// Position unimplemented
func (mti *xsense) Position(ctx context.Context, extra map[string]interface{}) (*geo.Point, float64, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return nil, 0, nil
}

// Properties
func (mti *xsense) Properties(ctx context.Context, extra map[string]interface{}) (*movementsensor.Properties, error) {
	mti.mu.Lock()
	defer mti.mu.Unlock()
	return &movementsensor.Properties{
		CompassHeadingSupported: true,
	}, nil
}

// Readings
func (mti *xsense) Readings(ctx context.Context, extra map[string]interface{}) (map[string]interface{}, error) {
	readings := make(map[string]interface{})
	return readings, nil
}

func newXsense(
	ctx context.Context,
	deps resource.Dependencies,
	conf resource.Config,
	logger golog.Logger,
) (movementsensor.MovementSensor, error) {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}

	options := slib.OpenOptions{
		PortName:        newConf.SerialPath,
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	if newConf.SerialBaudRate > 0 {
		options.BaudRate = uint(newConf.SerialBaudRate)
	} else {
		logger.Warnf(
			"no valid serial_baud_rate set, setting to default of %d, baud rate of wit imus are: %v", options.BaudRate, baudRateList,
		)
	}

	i := xsense{
		Named:  conf.ResourceName().AsNamed(),
		logger: logger,
		err:    movementsensor.NewLastError(1, 1),
	}
	logger.Debugf("initializing wit serial connection with parameters: %+v", options)
	i.port, err = slib.Open(options)
	if err != nil {
		return nil, err
	}

	return &i, nil
}
