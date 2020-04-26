#!/usr/bin/env python

import sys
import argparse
import json
import shlex
import subprocess

import requests

QUERY = """
resources[certname, type, title, parameters]{
  certname ~ "probe"
  and type = "Mon_check"
  and parameters.command ~ "nagios_graphite"
}
"""

FUNCTIONS = [
    "sum",
    "min",
    "max",
    "avg",
    "median",
    "95th",
    "99th",
    "999th",
    "nullcnt",
    "nullpct",
]


def make_optparse():
    from optparse import OptionParser, make_option

    hostname = make_option("-H", "--hostname", type="string", default=None)
    warning = make_option("-w", "--warning", type="string")
    critical = make_option("-c", "--critical", type="string")
    timeout = make_option("-t", "--timeout", type="int", default=0)
    verbosity = make_option("-v", "--verbose", action="count")
    username = make_option("--username", "-U", help="Username (HTTP Basic Auth)")
    password = make_option("--password", "-P", help="Password (HTTP Basic Auth)")

    name = make_option("--name", "-N", help="Metric name", default="metric")
    target = make_option("--target", "-M", help="Graphite target (series or query)")
    from_ = make_option("--from", "-F", help="Starting offset", default="1minute")

    until = make_option("--until", "-u", help="Ending offset", default="")

    func = make_option(
        "--algorithm", "-A", help="aggs", default="avg", choices=FUNCTIONS
    )

    http_timeout = make_option(
        "--http-timeout", "-o", help="HTTP request timeout", default=10, type=int
    )

    return OptionParser(
        option_list=[
            hostname,
            warning,
            critical,
            timeout,
            verbosity,
            username,
            password,
            name,
            target,
            from_,
            until,
            func,
            http_timeout,
        ]
    )


NG_PARSER = make_optparse()


def parse_args(args=None, parse=True):
    parser = argparse.ArgumentParser("Compare nagios_graphite and check-graphite")
    parser.add_argument(
        "--cg-exec", "-c", help="check-graphite exec", default="./check-graphite"
    )
    parser.add_argument(
        "--ng-exec", "-n", help="nagios_graphite exec", default="nagios-graphite"
    )
    parser.add_argument("--puppet", "-p", help="puppetdb server")
    parser.add_argument("--config", "-f", help="check configs")
    parser.add_argument("--dump", "-d", help="dump file")
    parser.add_argument(
        "--run", "-r", action="store_true", default=False, help="run the test"
    )
    parser.add_argument("--query", "-q", default=QUERY, help="puppetdb pql query")

    res = parser.parse_args(args) if parse else None
    return parser, res


def prune(d):
    return {k: v.strip() for k, v in d.items() if v and v.strip()}


def to_check_graphite(args):
    params = prune(
        {
            "-g": args.get("hostname"),
            "-w": args.get("warning"),
            "-c": args.get("critical"),
            "-a": args.get("algorithm"),
            "-n": args.get("name"),
            "-m": args.get("target"),
            "-f": args.get("from"),
            "-u": args.get("until")
        }
    )

    pl = [e for es in params.items() for e in es]
    return shlex.join(pl)


def extract_config(resource):
    cmd = resource.get("parameters", {}).get("command", "").strip()
    if not cmd or not cmd.startswith("nagios_graphite"):
        return

    argv = shlex.split(cmd)
    parsed = NG_PARSER.parse_args(argv)
    args = vars(parsed[0])
    if not args:
        return
    cg = to_check_graphite(args)
    return {"nagios_graphite": shlex.join(argv[1:]), "check-graphite": cg}


def load_puppet(puppet, query):
    if not puppet.startswith("http"):
        scheme = "https://"
    else:
        scheme = ""
    url = "{}{}/pdb/query/v4".format(scheme, puppet)
    payload = {"query": query}
    res = requests.post(url, json=payload)
    config = []
    for row in res.json():
        c = extract_config(row)
        if c:
            config.append(c)
    return config


def load_config(path):
    return json.load(open(path))

def run(cmd):
    proc = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    out, err = proc.communicate()
    return (proc.returncode, out+err)

def extract_perfdata(res):
    parts = res.split(b"|")
    sections = parts[1].split(b";")
    kv = sections[0].split(b"=")
    return float(kv[1])

def compare_results(ng, cg):
    diff = []
    if ng[0] != cg[0]:
        diff.append(f"ng returned {ng[0]}, but cg returned {cg[0]}")
    if ng[0] == cg[0] and ng[0] in [0, 1, 2]:
        ngpd = extract_perfdata(ng[1])
        cgpd = extract_perfdata(cg[1])
        try:
            chg = abs(ngpd - cgpd) / ngpd
            if chg > 0.03:
                diff.append(f"cg is {chg*100}% different than ng (ng:{ngpd}, cg:{cgpd})")
        except ZeroDivisionError:
            if cgpd > ngpd:
                diff.append(f"ng is 0 but cg is {cgpd}")

    return diff

def run_all(config, ng_exec, cg_exec):
    for conf in config:
        ng = f"{ng_exec} {conf['nagios_graphite']}"
        cg = f"{cg_exec} {conf['check-graphite']}"
        ng_res = run(ng)
        cg_res = run(cg)
        diff = compare_results(ng_res, cg_res)
        if diff:
            print(f"\nfor query: {cg}")
            for d in diff:
                print(f"* {d}")
            print("\n")

def main():
    _parser, args = parse_args()

    if not (args.puppet or args.config):
        print("must supply either puppet server or config file")

    if args.puppet:
        config = load_puppet(args.puppet, args.query)
    elif args.config:
        config = load_config(args.config)

    if args.dump:
        with open(args.dump) as fd:
            json.dump(fd, config)

    if args.run:
        run_all(config, args.ng_exec, args.cg_exec)

if __name__ == "__main__":
    main()
