#!/usr/bin/env python
"""Basic usage example and testing of pysma."""
# from time import sleep
import asyncio
import logging
import sys
import argparse
import aiohttp
import pysma
import time

# This module will work with Python 3.4+
# Python 3.4+ "@asyncio.coroutine" decorator
# Python 3.5+ uses "async def f()" syntax

_LOGGER = logging.getLogger(__name__)

args = None

METRICS = [
    {
        'name': '{}/sma/current_power_w',
        'key': pysma.KEY_CURRENT_POWER_W
    },
    {
        'name': '{}/sma/current_consumption_w',
        'key': pysma.KEY_CURRENT_CONSUMPTION_W
    },
    {
        'name': '{}/sma/total_yield_kwh',
        'key': pysma.KEY_TOTAL_YIELD_KWH
    },
    {
        'name': '{}/sma/total_consumption_kwh',
        'key': pysma.KEY_TOTAL_CONSUMPTION_KWH
    }
]

@asyncio.coroutine
def main(loop, password, ip, interval):
    """Main loop."""
    session = aiohttp.ClientSession(loop=loop)
    sma = pysma.SMA(session, ip, password=password,
                    group=pysma.GROUP_USER)
    yield from sma.new_session()

    while loop.jk_run:
        res = yield from sma.read([ m['key'] for m in METRICS ])
        t = time.time()
        for idx, metric in enumerate(METRICS):
            try:
                value = res[idx] if res[idx] != None else 0
                name = metric['name'].format(host)
                print("PUTVAL \"{}\" interval={} {}:{}".format(name, interval, int(t), value), flush=True)
            except TypeError:
                pass

        yield from asyncio.sleep(interval)

    yield from sma.close_session()
    yield from session.close()


if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.CRITICAL)

    parser = argparse.ArgumentParser(
        description='Poll an SMA webconnect instance for metrics, and expose them to collectd.')
    parser.add_argument(
        '--address', type=str, required=True, help='Network address of the Webconnect instance')
    parser.add_argument(
        '--password', type=str required=True, help='User password')
    parser.add_argument(
        '--host', type=str, required=True, help='Hostname to report collectd metric as')
    parser.add_argument(
        '--interval', type=int, default=10, help='Interval to emit metrics at')

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    try:
        setattr(loop, "jk_run", True)
        loop.run_until_complete(main(loop, password=password, ip=args.ip, host=args.host, interval=args.interval))
    finally:
        loop.close()
