package main

import (
	"context"
	"fmt"
	"os"

	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:  "route-p2p",
		Usage: "p2p router",
		Flags: []cli.Flag{
			// cli.IntFlag{
			// 	Name:   "grpc-port",
			// 	EnvVar: "GRPC_PORT",
			// 	Value:  12001,
			// },
			// cli.IntFlag{
			// 	Name:   "monitor-port",
			// 	EnvVar: "MONITOR_PORT",
			// },
			cli.StringFlag{
				Name:   "config",
				EnvVar: "P2PMQ_CONFIG",
				Value:  "/p2pmq/p2pmq.json",
			},
			cli.StringFlag{
				Name:   "loglevel",
				EnvVar: "P2PMQ_LOGLEVEL",
				Value:  "info",
			},
			cli.StringFlag{
				Name:   "libp2p-loglevel",
				EnvVar: "LIBP2P_LOGLEVEL",
				Value:  "info",
			},
		},
		Action: func(c *cli.Context) (err error) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			_ = logging.SetLogLevelRegex("p2p:.*", c.String("libp2p-loglevel"))
			_ = logging.SetLogLevelRegex("p2pmq.*", c.String("loglevel"))

			var d *core.Daemon
			if cfgPath := c.String("config"); len(cfgPath) > 0 {
				cfg, err := commons.ReadConfig(cfgPath)
				if err != nil {
					return err
				}
				d, err = core.NewDaemon(ctx, *cfg, nil, "node")
				if err != nil {
					return err
				}
				if err := d.Start(ctx); err != nil {
					return err
				}
				defer func() {
					_ = d.Close()
				}()
			}

			if d == nil {
				return fmt.Errorf("could not create daemon instance, please provide a config file")
			}

			<-ctx.Done()

			logging.Logger("p2pmq/cli").Info("closing node")

			return nil

			// svc := service.NewGrpc(ctx, c.String("name"))
			// defer svc.Close()

			// if monitorPort := c.Int("monitor-port"); monitorPort > 0 {
			// 	mux := http.NewServeMux()
			// 	monitoring.WithMetrics(mux)
			// 	monitoring.WithProfiling(mux)
			// 	monitoring.WithHealthCheck(mux, func() []error {
			// 		err := ctx.Err()
			// 		if err != nil {
			// 			return []error{err}
			// 		}
			// 		return nil
			// 	})
			// 	go func() {
			// 		err := http.ListenAndServe(fmt.Sprintf(":%d", monitorPort), mux)
			// 		if err != nil {
			// 			panic(err)
			// 		}
			// 	}()
			// }

			// s := svc.GrpcServer()
			// return service.ListenGrpc(s, c.Int("grpc-port"))
		},
		Commands: []cli.Command{},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}