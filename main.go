package main

import (
	"fmt"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/tristan-club/cascade/pkg/dbutil"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	"github.com/tristan-club/cascade/pkg/xrpc"
	"github.com/tristan-club/cascade/service"
	"github.com/tristan-club/kit/config"
	"github.com/tristan-club/kit/cron"
	tlog "github.com/tristan-club/kit/log"
	"github.com/xssnick/tonutils-go/address"
	"os"
	"time"
)

func main() {
	time.Local = time.FixedZone("UTC", 0)
	_ = address.Address{}

	xrpc.Init(xrpc.NameTaskMonitor, os.Getenv("TASK_MONITOR_SERVICE_ADDR"))
	if err := rdb.InitRedis(os.Getenv("REDIS_SERVICE_ADDR"), os.Getenv("REDIS_DB")); err != nil {
		panic(fmt.Errorf("init redis error: %s", err.Error()))
	}
	ms := migrate.Up
	if err := dbutil.InitDB(ms); err != nil {
		panic(fmt.Errorf("Init db error: %s ", err.Error()))
	}

	err := service.InitBot(os.Getenv("BOT_TOKEN"))
	if err != nil {
		panic(fmt.Errorf("init bot error: %s", err.Error()))
	}

	_ = cron.AddCronJob(service.DoWithdrawTxJob, "0 0 3 * *", config.EnvIsDev())

	tlog.Info().Msgf("start cascade success")
	c := make(chan bool, 2)
	<-c
}
