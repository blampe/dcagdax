package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gopkg.in/alecthomas/kingpin.v2"

	exchange "github.com/blampe/go-coinbase-exchange"
)

var (
	coin = kingpin.Flag(
		"coin",
		"Which coin you want to buy, BTC, LTC, ETH",
	).Required().String()

	every = registerGenerousDuration(kingpin.Flag(
		"every",
		"How often to make BTC purchases, e.g. 1h, 7d, 3w.",
	).Required())

	usd = kingpin.Flag(
		"usd",
		"How much USD to spend on each BTC purchase.",
	).Required().Float()

	until = registerDate(kingpin.Flag(
		"until",
		"Stop executing trades after this date, e.g. 2017-12-31.",
	))

	makeTrades = kingpin.Flag(
		"trade",
		"Actually execute trades.",
	).Bool()

	autoFund = kingpin.Flag(
		"autofund",
		"Automatically initiate ACH deposits.",
	).Bool()
)

func main() {
	kingpin.Version("0.1.0")
	kingpin.Parse()

	l, _ := zap.NewProduction()
	logger := l.Sugar()
	defer logger.Sync()

	secret := os.Getenv("GDAX_SECRET")
	key := os.Getenv("GDAX_KEY")
	passphrase := os.Getenv("GDAX_PASSPHRASE")

	if secret == "" {
		logger.Warn("GDAX_SECRET environment variable is required")
		os.Exit(1)
	} else {
		os.Setenv("COINBASE_SECRET", secret)
	}
	if key == "" {
		logger.Warn("GDAX_KEY environment variable is required")
	} else {
		os.Setenv("COINBASE_KEY", key)
	}
	if passphrase == "" {
		logger.Warn("GDAX_PASSPHRASE environment variable is required")
	} else {
		os.Setenv("COINBASE_PASSPHRASE", key)
	}

	client := exchange.NewClient(secret, key, passphrase)

	schedule, err := newGdaxSchedule(
		client,
		logger,
		!*makeTrades,
		*autoFund,
		*usd,
		*every,
		*until,
		*coin,
	)

	if err != nil {
		logger.Warn(err.Error())
		os.Exit(1)
	}

	if err := schedule.Sync(); err != nil {
		logger.Warn(err.Error())
	}
}

type generousDuration time.Duration

func registerGenerousDuration(s kingpin.Settings) (target *time.Duration) {
	target = new(time.Duration)
	s.SetValue((*generousDuration)(target))
	return
}

func (d *generousDuration) Set(value string) error {
	durationRegex := regexp.MustCompile(`^(?P<value>\d+)(?P<unit>[hdw])$`)

	if !durationRegex.MatchString(value) {
		return fmt.Errorf("--every misformatted")
	}

	matches := durationRegex.FindStringSubmatch(value)

	hours, _ := strconv.ParseInt(matches[1], 10, 64)
	unit := matches[2]

	switch unit {
	case "d":
		hours *= 24
	case "w":
		hours *= 24 * 7
	}

	duration := time.Duration(hours * int64(time.Hour))

	*d = (generousDuration)(duration)

	return nil
}

func (d *generousDuration) String() string {
	return (*time.Duration)(d).String()
}

type date time.Time

func (d *date) Set(value string) error {
	t, err := time.Parse("2006-01-02", value)

	if err != nil {
		return err
	}

	*d = (date)(t)

	return nil
}

func (d *date) String() string {
	return (*time.Time)(d).String()
}

func registerDate(s kingpin.Settings) (target *time.Time) {
	target = &time.Time{}
	s.SetValue((*date)(target))
	return target
}
