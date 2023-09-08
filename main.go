package main

import (
	"context"

	"github.com/edaniels/golog"
	xsense "github.com/viam-labs/mti-xsense/xsense"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/utils"
)

func main() {
	utils.ContextualMain(mainWithArgs, golog.NewDevelopmentLogger("mti-xsense"))
}

func mainWithArgs(ctx context.Context, args []string, logger golog.Logger) error {
	imu, err := module.NewModuleFromArgs(ctx, logger)
	if err != nil {
		return err
	}

	imu.AddModelFromRegistry(ctx, sensor.Subtype, xsense.Model)

	err = imu.Start(ctx)
	defer imu.Close(ctx)
	if err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
