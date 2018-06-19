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

from pprint import pprint

# This module will work with Python 3.4+
# Python 3.4+ "@asyncio.coroutine" decorator
# Python 3.5+ uses "async def f()" syntax

_LOGGER = logging.getLogger(__name__)

args = None

METRICS = [
    {
        'name': '***REMOVED***/sma/current_power_w',
        'key': pysma.KEY_CURRENT_POWER_W
    },
    {
        'name': '***REMOVED***/sma/current_consumption_w',
        'key': pysma.KEY_CURRENT_CONSUMPTION_W
    },
    {
        'name': '***REMOVED***/sma/total_yield_kwh',
        'key': pysma.KEY_TOTAL_YIELD_KWH
    },
    {
        'name': '***REMOVED***/sma/total_consumption_kwh',
        'key': pysma.KEY_TOTAL_CONSUMPTION_KWH
    }
]

@asyncio.coroutine
def main(loop, password, ip, interval):
    """Main loop."""
    session = aiohttp.ClientSession(loop=loop)
    sma = pysma.SMA(session, ip, password=password,
                    group=pysma.GROUP_USER)
#    pprint(vars(sma))
    yield from sma.new_session()
#    _LOGGER.info("NEW SID: %s", sma._sma_sid)

#    import code; code.interact(local=dict(globals(), **locals()))

    while loop.jk_run:
        res = yield from sma.read([ m['key'] for m in METRICS ])
        t = time.time()
        for idx, metric in enumerate(METRICS):
            try:
                value = res[idx] if res[idx] != None else 0
                print("PUTVAL \"{}\" interval={} {}:{}".format(metric['name'], interval, int(t), value), flush=True)
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
        'ip', type=str, help='IP address of the Webconnect module')
#    parser.add_argument(
#        'password', help='User password')
    password = '***REMOVED***'

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    try:
        setattr(loop, "jk_run", True)
        loop.run_until_complete(main(loop, password=password, ip=args.ip, interval=10))
    finally:
        loop.close()
