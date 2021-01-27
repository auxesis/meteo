#!/usr/bin/env python
"""Collectd read plugin for SMA Sunny Boy inverters."""
import atexit
import asyncio
import async_timeout
import logging
import sys
import argparse
import aiohttp
import pysma
import time
import traceback

_LOGGER = logging.getLogger(__name__)

args = None


def sensor(name):
    try:
        return next(s for s in pysma.SENSORS if s.name == name)
    except StopIteration:
        return None


METRICS = [
    {"name": "{}/sma/current_power_w", "sensor": sensor("current_power")},
    {"name": "{}/sma/current_consumption_w", "sensor": sensor("current_consumption")},
    {"name": "{}/sma/total_yield_kwh", "sensor": sensor("total_yield")},
    {"name": "{}/sma/total_consumption_kwh", "sensor": sensor("total_consumption")},
]


def putval(res):
    t = time.time()
    for idx, metric in enumerate(METRICS):
        try:
            value = res[idx] if res[idx] != None else 0
            name = metric["name"].format(host)
            print(
                'PUTVAL "{}" interval={} {}:{}'.format(name, interval, int(t), value),
                flush=True,
            )
        except TypeError:
            pass


async def cleanup(sma, session):
    print("Cleaning up sessions...")
    await sma.close_session()
    await session.close()


async def main(loop, address, password, host, interval):
    """Main loop."""
    session = aiohttp.ClientSession(loop=loop)
    sma = pysma.SMA(session, address, password=password, group="user")
    await sma.new_session()

    atexit.register(await cleanup, sma, session)

    while True:
        try:
            with async_timeout.timeout(20):
                res = await sma.read([m["sensor"] for m in METRICS])
            putval(res)
        except asyncio.TimeoutError:
            _LOGGER.error(traceback.format_exc())
            _LOGGER.error("Timeout when talking to SMA inverter")
            sys.exit(3)

        await asyncio.sleep(interval)


if __name__ == "__main__":
    logging.basicConfig(stream=sys.stdout, level=logging.DEBUG)

    parser = argparse.ArgumentParser(
        description="Poll an SMA webconnect instance for metrics, and expose them to collectd."
    )
    parser.add_argument(
        "--address",
        type=str,
        required=True,
        help="Network address of the Webconnect instance",
    )
    parser.add_argument(
        "--password",
        type=str,
        required=True,
        help="User password for Webconnect instance",
    )
    parser.add_argument(
        "--host", type=str, required=True, help="Hostname to report collectd metrics as"
    )
    parser.add_argument(
        "--interval", type=int, default=10, help="Interval to emit metrics at"
    )

    args = parser.parse_args()

    loop = asyncio.get_event_loop()
    try:
        loop.run_until_complete(
            main(
                loop,
                password=args.password,
                address=args.address,
                host=args.host,
                interval=args.interval,
            )
        )
    except KeyboardInterrupt:
        print("Caught C-c – exiting.")
        loop.close()
        sys.exit(0)
    except Exception:
        loop.close()
        sys.exit(2)
    finally:
        loop.close()
        sys.exit(0)
