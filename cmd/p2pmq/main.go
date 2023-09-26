package main

import (
	"context"
	"fmt"
	"os"
	"time"

	grpcapi "github.com/amirylm/p2pmq/api/grpc"
	"github.com/amirylm/p2pmq/commons"
	"github.com/amirylm/p2pmq/core"
	"github.com/amirylm/p2pmq/core/gossip"
	logging "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name: "p2pmq",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:   "grpc-port",
				EnvVar: "GRPC_PORT",
				Value:  8001,
			},
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

			lggr := logging.Logger("p2pmq/cli")
			cfgPath := c.String("config")
			if len(cfgPath) == 0 {
				return fmt.Errorf("missing config file")
			}
			cfg, err := commons.ReadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("failed to read config file: %w", err)
			}
			rlggr := lggr.Named("msgRouter")
			msgRouter := core.NewMsgRouter(cfg.MsgRouterQueueSize, cfg.MsgRouterWorkers, func(mw *core.MsgWrapper[error]) {
				rlggr.Debugw("Got message", "from", mw.Peer, "msg", mw.Msg)
			}, gossip.MsgIDSha256(20))
			valRouter := core.NewMsgRouter(cfg.MsgRouterQueueSize, cfg.MsgRouterWorkers, func(mw *core.MsgWrapper[pubsub.ValidationResult]) {
				rlggr.Debugw("Validating message", "from", mw.Peer, "msg", mw.Msg)
				mw.Result = pubsub.ValidationAccept
			}, gossip.MsgIDSha256(20))

			ctrl, err := core.NewController(ctx, *cfg, msgRouter, valRouter, "node")
			if err != nil {
				return err
			}

			control, msgR, valR := grpcapi.NewServices(ctrl, 128)
			ctrl.RefreshRouters(func(mw *core.MsgWrapper[error]) {
				rlggr.Debugw("Got message", "from", mw.Peer, "msg", mw.Msg)
				if err := msgR.Push(mw); err != nil {
					rlggr.Debugw("Failed to push incoming message", "from", mw.Peer, "msg", mw.Msg)
				}
			}, func(mw *core.MsgWrapper[pubsub.ValidationResult]) {
				rlggr.Debugw("Validating message", "from", mw.Peer, "msg", mw.Msg)
				ctx, cancel := context.WithTimeout(ctx, time.Second*5)
				defer cancel()

				mw.Result = valR.PushWait(ctx, mw)
			})
			srv := grpcapi.NewGrpcServer(control, msgR, valR)

			ctrl.Start(ctx)
			defer ctrl.Close()

			// <-time.After(time.Second * 10)

			// if cfg.Pubsub != nil {
			// 	if err := ctrl.Subscribe(ctx, "test-1"); err != nil {
			// 		lggr.Errorw("could not subscribe to topic", "topic", "test-1", "err", err)
			// 	}
			// 	for i := 0; i < 10; i++ {
			// 		<-time.After(time.Second * 5)
			// 		if err := ctrl.Publish(ctx, "test-1", []byte(fmt.Sprintf("test-data-%d-%s", i, ctrl.ID()))); err != nil {
			// 			lggr.Errorw("could not subscribe to topic", "topic", "test-1", "err", err)
			// 		}
			// 	}
			// }

			return grpcapi.ListenGrpc(srv, c.Int("grpc-port"))

		},
		Commands: []cli.Command{},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
