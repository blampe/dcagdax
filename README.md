你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
# DCA GDAX

Automated dollar cost averaging for BTC, LTC, BCH and ETH on GDAX.

## Setup

If you only have a Coinbase account you'll need to also sign into
[GDAX](gdax.com). Make sure you have a bank account linked to one of these for
ACH transfers.

Procure a GDAX API key for yourself by visiting
[www.gdax.com/settings/api](https://www.gdax.com/settings/api). **Do not share
this API key with third parties!**

## Usage

Build the binary:

```
$ go build ./
```

Then run it:

```
./dcagdax --help
usage: dcagdax --every=EVERY [<flags>]

Flags:
  --help         Show context-sensitive help (also try --help-long and
                 --help-man).
  --coin=BTC     Which coin you want to buy: BTC, LTC, BCH or ETH (default 'BTC').
  --every=EVERY  How often to make purchases, e.g. 1h, 7d, 3w.
  --usd=USD      How much USD to spend on each purchase. If unspecified, the
                 minimum purchase amount allowed will be used.
  --until=UNTIL  Stop executing trades after this date, e.g. 2017-12-31.
  --trade        Actually execute trades.
  --autofund     Automatically initiate ACH deposits.
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

**Q:** Which coins can I purchase?

**A:** We support all of GDAX's products: BTC, LTC, BCH and ETH.

**Q:** Can I buy you a beer?

**A:** BTC `13N2g3MedU2mc6TzQWEE7kBoTCj6Ajyzpw`
