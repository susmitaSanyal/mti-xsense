package xsense

import (
	"context"
	"io"
	"sync"

	"github.com/edaniels/golog"
	"github.com/golang/geo/r3"
	slib "github.com/jacobsa/go-serial/serial"
	"github.com/pkg/errors"
	"go.viam.com/utils"

	"go.viam.com/rdk/components/movementsensor"
	"go.viam.com/rdk/resource"
	rutils "go.viam.com/rdk/utils"
)

var Model = resource.NewModel("viam", "sensor", "mti-xsense-200")
var baudRateList = []uint{115200, 9600, 0}

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
