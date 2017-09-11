# DCA GDAX

Automated dollar cost averaging for BTC on GDAX.

## Setup

If you only have a Coinbase account you'll need to also sign into
[GDAX](gdax.com). Make sure you have a bank account linked to one of these for
ACH transfers.

Procure a GDAX API key for yourself by visiting
[www.gdax.com/settings/api](https://www.gdax.com/settings/api). **Do not share
this API key with third parties!**

Install dependencies with [glide](https://github.com/Masterminds/glide):

```
$ glide i
```

You should now be able to build the binary without seeing any errors:

```
$ go build ./
```

## Usage

```
$ ./dcagdax --help
usage: dcagdax --every=EVERY --usd=USD [<flags>]

Flags:
  --help         Show context-sensitive help (also try --help-long and --help-man).
  --every=EVERY  How often to make BTC purchases, e.g. 1h, 7d, 3w.
  --usd=USD      How much USD to spend on each BTC purchase.
  --until=UNTIL  Stop executing trades after this date, e.g. 2017-12-31.
  --trade        Actually execute trades.
  --version      Show application version.
```

Run the `dcagdax` binary with an environment containing your API credentials:
```
$ GDAX_SECRET=secret \
  GDAX_KEY=key \
  GDAX_PASSPHRASE=pass \
  ./dcagdax --help
```

Be aware that if you set your purchase amount near 0.01 BTC (the minimum trade
amount) then an upswing in price might prevent you from trading.

## FAQ

**Q:** Why do I not see any trading activity from the bot?

**A:** If you have other BTC trades on your account, the bot will detect that as a
cost-averaged purchase and wait until the next purchase window. This is for
people who want to "set it and forget it," not day traders!

**Q:** Why would I use this instead of Coinbase's recurring purchase feature?

**A:** The [fees on recurring
purchases](https://support.coinbase.com/customer/portal/articles/2109597)
(currently a minimum of $2.99 per purchase!) can add up quickly. This
side-steps those costs by automating free ACH deposits into your exchange
account & submitting market orders to exchange with BTC.

**Q:** How should I deploy this?

**A:** You could run this as a periodic cronjob on your workstation or in the
cloud. Just be sure your API key & secret are not made available to anyone else
as part of your deployment!

**Q:** Can this auto-withdraw coins into a cold wallet?

**A:** Not currently, but pull requests are welcome!

**Q:** Can I buy you a beer?

**A:** BTC `13N2g3MedU2mc6TzQWEE7kBoTCj6Ajyzpw`
