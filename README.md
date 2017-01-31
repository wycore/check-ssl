# check-ssl
[![GitHub version](https://badge.fury.io/gh/wywygmbh%2Fcheck-ssl.svg)](https://badge.fury.io/gh/wywygmbh%2Fcheck-ssl)
[![Build Status](https://travis-ci.org/wywygmbh/check-ssl.svg?branch=master)](https://travis-ci.org/wywygmbh/check-ssl)


Monitor SSL certificate validity for records with multiple IPs.

Compatible with Icinga, Nagios, Sensu, ... It uses the common exit codes.

## Motivation

We have several domains which are using DNS RR for loadbalancing/availability. 

Such domains are especially sensitive, when SSL certificates are renewed. Some of them can easily be missed or deployment fails. It's often hard to discover, which particular service is misconfigured.

So we've created this little tool to fill this gap. It will resolve all IPs belonging to a certain domain, check every one of them, validates that the certificate chains are correct and that they are still valid for some time.

It will handle both IPv4 and IPv6. Missing IPv6 connectivity will be reported in the result.

## Example

    $ check-ssl -host analytics.wywy.com
    INFO 52.222.157.227  - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.169  - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.182  - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.140  - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.73   - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.18   - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 59s) 
    INFO 52.222.157.90   - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 58s) 
    INFO 52.222.157.121  - *.wywy.com valid until 2018-11-16 23:59:59 +0000 UTC (654d 23h 56m 58s)
 
## Usage

    Usage of ./check-ssl:
      -V	print version and exit
      -c uint
            critical validity in days (default 14)
      -connection-timeout duration
            timeout connection - see: https://golang.org/pkg/time/#ParseDuration (default 30s)
      -host string
            the domain name of the host to check
      -lookup-timeout duration
            timeout for DNS lookups - see: https://golang.org/pkg/time/#ParseDuration (default 10s)
      -w uint
            warning validity in days (default 30)

## License

Copyright 2017 wywy GmbH

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

`simple_text_formatter.go` is based on the logrus project and licensed under the MIT License. See LICENSE-MIT for details. 

This code is being actively maintained by some fellow engineers at [wywy GmbH](http://wywy.com/).