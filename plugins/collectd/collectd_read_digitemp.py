#!/usr/bin/env python

import asyncio
import logging
import sys
import aiohttp
import time
import requests
import argparse
import os
# This module will work with Python 3.4+
# Python 3.4+ "@asyncio.coroutine" decorator
# Python 3.5+ uses "async def f()" syntax

_LOGGER = logging.getLogger(__name__)

METRICS = [
    {
        'name': '{}/{}/temperature-air',
        'key': 'air_temp'
    },
]

@asyncio.coroutine
def digitemp_read_temperature(keys):
    try:
        cmd = 'digitemp_DS9097 -c /etc/digitemp.conf -q -t 0 -s /dev/ttyUSB0 -r 1000 -o "%.2C"'
        for line in os.popen(cmd).readlines():
            return [ line.strip() ]
    except:
        return []

@asyncio.coroutine
def main(loop, host, plugin, interval):
    """Main loop."""
    while loop.jk_run:
        res = yield from digitemp_read_temperature([ m['key'] for m in METRICS ])
        t = time.time()
        for idx, metric in enumerate(METRICS):
            value = res[idx] if res[idx] != None else 0
            name = metric['name'].format(host, plugin)
            print("PUTVAL \"{}\" interval={} {}:{}".format(name, interval, int(t), value), flush=True)
        yield from asyncio.sleep(interval)

if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.CRITICAL)

    parser = argparse.ArgumentParser(
        description='Turn weather readings into collectd metrics')
    parser.add_argument(
        '--host', type=str, required=True, help='Hostname to report collectd metric as')
    parser.add_argument(
        '--plugin', type=str, required=True, help='Plugin to report collectd metric as')
    parser.add_argument(
        '--interval', type=int, default=10, help='Interval to emit metrics at')

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    try:
        setattr(loop, "jk_run", True)
        loop.run_until_complete(main(loop, host=args.host, plugin=args.plugin, interval=args.interval))
    except KeyboardInterrupt:
        setattr(loop, "jk_run", False)
        loop.run_forever()
        _LOGGER.info('Done (Ctrl-C)')
