package main

import (
	"errors"
	"fmt"
	"math"
	"time"

	exchange "github.com/blampe/go-coinbase-exchange"
	"go.uber.org/zap"
)

var skippedForDebug = errors.New("Skipping because trades are not enabled")

type gdaxSchedule struct {
	logger *zap.SugaredLogger
	client *exchange.Client
	debug  bool

	usd      float64
	every    time.Duration
	until    time.Time
	autoFund bool
	coin     string
}

func newGdaxSchedule(
	c *exchange.Client,
	l *zap.SugaredLogger,
	debug bool,
	autoFund bool,
	usd float64,
	every time.Duration,
	until time.Time,
	coin string,
) (*gdaxSchedule, error) {
	schedule := gdaxSchedule{
		logger: l,
		client: c,
		debug:  debug,

		usd:      usd,
		every:    every,
		until:    until,
		autoFund: autoFund,
		coin:     coin,
	}

	minimum, err := schedule.minimumUSDPurchase()

	if err != nil {
		return nil, err
	}

	if schedule.usd < minimum {
		return nil, errors.New(fmt.Sprintf(
			"GDAX's minimum %s trade amount is $%.02f, but you're trying to purchase $%f",
			schedule.coin, minimum, schedule.usd,
		))
	}

	return &schedule, nil
}

// Sync initiates trades & funding with a DCA strategy.
func (s *gdaxSchedule) Sync() error {

	now := time.Now()

	until := s.until
	if until.IsZero() {
		until = time.Now()
	}

	if now.After(until) {
		return errors.New("Deadline has passed, not taking any action")
	}

	s.logger.Infow("Dollar cost averaging",
		"USD", s.usd,
		"every", every,
		"until", until.String(),
	)

	if time, err := s.timeToPurchase(); err != nil {
		return err
	} else if !time {
		return errors.New("Detected a recent purchase, waiting for next purchase window")
	}

	if funded, err := s.sufficientUsdAvailable(); err != nil {
		return err
	} else if !funded {
		needed, err := s.additionalUsdNeeded()
		if err != nil {
			return err
		}

		if needed == 0 {
			return errors.New("Not enough available funds, wait for transfers to settle")
		}

		if needed > 0 {
			s.logger.Infow(
				"Insufficient funds",
				"needed", needed,
			)
			if s.autoFund {
				s.logger.Infow(
					"TODO: Creating a transfer request for $%.02f",
					"needed", needed,
				)
				s.makeDeposit(needed)
			}
		}
		return nil
	}

	s.logger.Infow(
		"Placing an order",
		"coin", s.coin,
		"purchaseCurrency", "USD",
		"purchaseAmount", s.usd,
	)

	productId := s.coin + "-" + "USD"
	if err := s.makePurchase(productId); err != nil {
		s.logger.Warn(err)
	}

	return nil
}

func (s *gdaxSchedule) minimumUSDPurchase() (float64, error) {
	productId := s.coin + "-" + "USD"
	ticker, err := s.client.GetTicker(productId)

	if err != nil {
		return 0, err
	}

	products, err := s.client.GetProducts()

	if err != nil {
		return 0, err
	}

	for _, p := range products {
		if p.BaseCurrency == s.coin {
			return p.BaseMinSize * ticker.Price, nil
		}
	}

	return 0, errors.New(productId + " not found")
}

func (s *gdaxSchedule) timeToPurchase() (bool, error) {
	timeSinceLastPurchase, err := s.timeSinceLastPurchase()

	if err != nil {
		return false, err
	}

	if timeSinceLastPurchase.Seconds() < s.every.Seconds() {
		// We purchased something recently, so hang tight.
		return false, nil
	}

	return true, nil
}

func (s *gdaxSchedule) sufficientUsdAvailable() (bool, error) {
	usdAccount, err := s.accountFor("USD")

	if err != nil {
		return false, err
	}

	return (usdAccount.Available >= s.usd), nil
}

func (s *gdaxSchedule) additionalUsdNeeded() (float64, error) {
	if funded, err := s.sufficientUsdAvailable(); err != nil {
		return 0, err
	} else if funded {
		return 0, nil
	}

	usdAccount, err := s.accountFor("USD")
	if err != nil {
		return 0, nil
	}

	dollarsNeeded := s.usd - usdAccount.Available
	if dollarsNeeded < 0 {
		return 0, errors.New("Invalid account balance")
	}

	// Dang, we don't have enough funds. Let's see if money is on the way.
	var transfers []exchange.Transfer
	cursor := s.client.ListAccountTransfers(usdAccount.Id)

	dollarsInbound := 0.0

	for cursor.HasMore {
		if err := cursor.NextPage(&transfers); err != nil {
			return 0, err
		}

		for _, t := range transfers {
			unprocessed := (t.ProcessedAt.Time() == time.Time{})
			notCanceled := (t.CanceledAt.Time() == time.Time{})

			// This transfer is stil pending, so count it.
			if unprocessed && notCanceled {
				dollarsInbound += t.Amount
			}
		}
	}

	// If our incoming transfers don't cover our purchase need then we'll need
	// to cover that with an additional deposit.
	return math.Max(dollarsNeeded-dollarsInbound, 0), nil
}

func (s *gdaxSchedule) timeSinceLastPurchase() (time.Duration, error) {
	var transactions []exchange.LedgerEntry
	account, err := s.accountFor(s.coin)
	if err != nil {
		return 0, err
	}
	cursor := s.client.ListAccountLedger(account.Id)

	lastTransactionTime := time.Time{}
	now := time.Now()

	for cursor.HasMore {
		if err := cursor.NextPage(&transactions); err != nil {
			return 0, err

		}

		// Consider trade transactions
		for _, t := range transactions {
			if t.CreatedAt.Time().After(lastTransactionTime) && t.Type == "match" {
				lastTransactionTime = t.CreatedAt.Time()
			}
		}
	}

	return now.Sub(lastTransactionTime), nil
}

func (s *gdaxSchedule) makePurchase(productId string) error {
	if s.debug {
		return skippedForDebug
	}

	order, err := s.client.CreateOrder(
		&exchange.Order{
			ProductId: productId,
			Type:      "market",
			Side:      "buy",
			Funds:     s.usd,
		},
	)

	if err != nil {
		return err
	}

	s.logger.Infow(
		"Placed order",
		"orderId", order.Id,
	)

	return nil
}

func (s *gdaxSchedule) makeDeposit(amount float64) error {
	// TODO: Initiate funding for this amount. Need to add
	// /deposits/payment-method support to client and
	// client.CreateTransfer(...)
	return skippedForDebug
}

func (s *gdaxSchedule) accountFor(currencyCode string) (*exchange.Account, error) {
	accounts, err := s.client.GetAccounts()
	if err != nil {
		return nil, err
	}

	for _, a := range accounts {
		if a.Currency == currencyCode {
			return &a, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("No %s wallet on this account", currencyCode))
}
