#!/usr/bin/env python

import asyncio
import logging
import sys
import aiohttp
import time
import requests
import argparse
import traceback

# This module will work with Python 3.4+
# Python 3.4+ "@asyncio.coroutine" decorator
# Python 3.5+ uses "async def f()" syntax

_LOGGER = logging.getLogger(__name__)

METRICS = [
    {
        'name': '{}/station/temperature-air',
        'key': 'air_temp'
    },
    {
        'name': '{}/station/temperature-apparent',
        'key': 'apparent_t'
    }
]

@asyncio.coroutine
def bom_read(area_id, station_id, keys):
    url = "http://www.bom.gov.au/fwo/{}/{}.{}.json".format(area_id, area_id, station_id)
    try:
        r = requests.get(url, timeout=10)
    except:
        _LOGGER.error(traceback.format_exc())
        _LOGGER.error("Failed when making GET request to URL {}".format(url))
        sys.exit(1)
    try:
        sample = r.json()['observations']['data'][0]
        return [ sample[k] for k in keys ]
    except:
        return []

@asyncio.coroutine
def main(loop, host, area_id, station_id, interval):
    """Main loop."""
    while loop.jk_run:
        res = yield from bom_read(area_id, station_id, [ m['key'] for m in METRICS ])
        t = time.time()
        for idx, metric in enumerate(METRICS):
            value = res[idx] if res[idx] != None else 0
            name = metric['name'].format(host)
            print("PUTVAL \"{}\" interval={} {}:{}".format(name, interval, int(t), value), flush=True)
        yield from asyncio.sleep(interval)

if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.INFO)

    parser = argparse.ArgumentParser(
        description='Turn weather readings into collectd metrics')
    parser.add_argument(
        '--host', type=str, required=True, help='Hostname to report collectd metric as')
    parser.add_argument(
        '--area-id', type=str, required=True, help='Area id of station')
    parser.add_argument(
        '--station-id', type=str, required=True, help='Station id of station')
    parser.add_argument(
        '--interval', type=int, default=60, help='Interval to emit metrics at')

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    try:
        setattr(loop, "jk_run", True)
        loop.run_until_complete(main(loop, host=args.host,
            area_id=args.area_id, station_id=args.station_id, interval=args.interval))
    except KeyboardInterrupt:
        setattr(loop, "jk_run", False)
        loop.run_forever()
        _LOGGER.info('Done (Ctrl-C)')
    except Exception:
        _LOGGER.error(traceback.format_exc())
        _LOGGER.error('Unhandled error. Exiting.')
        sys.exit(1)
