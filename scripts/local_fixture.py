#!/usr/bin/env python3
"""Owned rootless Ethernet fixture. Smoke and dry-run never certify L1/L2."""
from __future__ import annotations

import argparse
import ctypes
import collections
import fcntl
import hashlib
import json
import math
import os
from pathlib import Path
import resource
import signal
import socket
import sqlite3
import stat
import subprocess
import sys
import time
import uuid
from datetime import datetime, timezone
import urllib.request

if __package__:
    from .slo_ingress import frame
else:
    from slo_ingress import frame

NS = 10**9
GIB = 1024**3
BUDGET = 4 * GIB
# All children inherit this hard per-file limit. The fixed runner has fewer than
# 32 files/run, leaving an entire GiB reservation including DB sidecars/logs.
FILE_LIMIT = 32 * 1024**2
RUN_RESERVE = GIB
MAX_FILES = 32
LINK = dict(src_mac='02:00:00:00:00:01', dst_mac='02:00:00:00:00:02',
            src_ip='192.0.2.1', dst_ip='192.0.2.2', src_port=19001, dst_port=19002)
REPO = Path(__file__).resolve().parent.parent


def encode(value):
    return (json.dumps(value, sort_keys=True, separators=(',', ':'), allow_nan=False)+'\n').encode()


def save(path, value):
    with path.open('xb') as stream:
        stream.write(encode(value))
        stream.flush()
        os.fsync(stream.fileno())


def digest(path):
    h = hashlib.sha256()
    with path.open('rb') as stream:
        for block in iter(lambda: stream.read(1024**2), b''):
            h.update(block)
    return h.hexdigest()


def rules():
    return {'rules': [dict(id=f'local-{i:03}', name=f'Local oracle {i}',
                          type='payload_match', severity='medium', enabled=True,
                          config={'keywords': [f'LOCAL_POS_{i:03}_END']},
                          mitre_techniques=[]) for i in range(100)]}


def schedule(smoke=False):
    """Integer offsets avoid accumulated floating-point schedule drift."""
    sequence = 0
    for second in range(2 if smoke else 600):
        rate = 10 if smoke else 500 if 180 <= second < 240 or 480 <= second < 540 else 100
        for index in range(rate):
            sequence += 1
            yield sequence, second*NS + index*NS//rate, second % 100 if index == 0 else None


def packet(run_id, sequence, positive):
    packet_id = f'p{sequence:06}'
    payload = b'' if positive is None else f'LOCAL_POS_{positive:03}_END'.encode()
    # slo_ingress adds the marker; pad payload to exact Ethernet length 512.
    marker = f'NSLO1 {run_id} {packet_id}\n'.encode()
    payload = payload + b'.'*(512-42-len(marker)-len(payload))
    raw = frame(run_id, packet_id, payload, LINK, sequence)
    expected = [] if positive is None else [dict(rule_id=f'local-{positive:03}',
        event_id='slo_'+hashlib.sha256((packet_id+'\0'+f'local-{positive:03}').encode()).hexdigest())]
    return packet_id, raw, expected


def frozen(run_id, smoke=False):
    if not isinstance(run_id, str) or not 1 <= len(run_id) <= 40 or any(
            c not in 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_' for c in run_id):
        raise ValueError('run ID must contain 1..40 ASCII letters, digits, hyphens or underscores')
    count = expected = 0
    h = hashlib.sha256()
    for sequence, offset, positive in schedule(smoke):
        pid, raw, events = packet(run_id, sequence, positive)
        h.update(encode(dict(packet_id=pid, scheduled_ns=offset,
                             frame_sha256=hashlib.sha256(raw).hexdigest(), expected_alerts=events)))
        count += 1
        expected += len(events)
    return dict(schema_version=1, contract_id='local-v1', run_id=run_id,
                mode='smoke' if smoke else 'L1-fixture', duration_seconds=2 if smoke else 600,
                drain_seconds=3 if smoke else 30, planned_packets=count,
                expected_events=expected, frame_bytes=512, rules_sha256=hashlib.sha256(encode(rules())).hexdigest(),
                fixture_sha256=h.hexdigest(), schedule={'steady_pps':10 if smoke else 100,
                'burst_pps':None if smoke else 500, 'burst_seconds':[] if smoke else [[180,240],[480,540]]},
                sut={'cpu_count':4, 'memory_bytes':3*GIB}, generator={'cpu_count':2,'memory_bytes':GIB},
                evidence_budget_bytes=BUDGET, run_reservation_bytes=RUN_RESERVE,
                child_file_limit_bytes=FILE_LIMIT, max_files=MAX_FILES,
                execution_complete=False, slo_compliance_asserted=False)


def freeze_files(run, run_id, smoke=False):
    declaration = frozen(run_id,smoke)
    save(run/'frozen.json',declaration)
    save(run/'rules.json',rules())
    # Closed before SUT startup: planned identities are independent of successful
    # submission and engine output, and survive a generator dying mid-send.
    with (run/'oracle.jsonl').open('xb') as oracle:
        for sequence, offset, positive in schedule(smoke):
            pid, raw, expected = packet(run_id,sequence,positive)
            oracle.write(encode(dict(packet_id=pid,scheduled_ns=offset,
                frame_sha256=hashlib.sha256(raw).hexdigest(),expected_alerts=expected)))
        oracle.flush()
        os.fsync(oracle.fileno())
    if digest(run/'oracle.jsonl') != declaration['fixture_sha256']:
        raise RuntimeError('frozen oracle hash differs')
    return declaration


def inventory(root):
    total = count = 0
    for base, dirs, files in os.walk(root, followlinks=False):
        for name in dirs + files:
            meta = (Path(base)/name).lstat()
            if stat.S_ISLNK(meta.st_mode):
                raise ValueError('evidence tree must not contain symlinks')
            if stat.S_ISREG(meta.st_mode):
                if meta.st_nlink != 1:
                    raise ValueError('evidence tree must not contain hard links')
                total += meta.st_size
                count += 1
            elif not stat.S_ISDIR(meta.st_mode) and not stat.S_ISSOCK(meta.st_mode):
                raise ValueError('unsupported evidence object')
    return total, count


def admit(root, free=None):
    if root.is_symlink() or not root.is_dir() or root.stat().st_uid != os.getuid():
        raise ValueError('evidence root must be an existing owned directory')
    if root.stat().st_mode & 0o077:
        raise ValueError('evidence root must be private (0700)')
    used, _ = inventory(root)
    if used + RUN_RESERVE > BUDGET:
        raise ValueError('cumulative 4GiB budget cannot reserve another run')
    space = os.statvfs(root)
    available = space.f_bavail*space.f_frsize if free is None else free
    if available < 8*GIB:
        raise ValueError('at least 8GiB free disk required')
    return used


class Groups:
    """Only children of this fresh delegated systemd scope are manipulated."""
    def __init__(self):
        relative = Path('/proc/self/cgroup').read_text().strip().split('::',1)[1]
        self.root = Path('/sys/fs/cgroup'+relative)
        if not self.root.name.startswith('run-') or not self.root.name.endswith('.scope'):
            raise ValueError('requires a fresh delegated user-systemd scope')
        (self.root/'control').mkdir()
        (self.root/'control/cgroup.procs').write_text(str(os.getpid()))
        (self.root/'cgroup.subtree_control').write_text('+memory +cpu')
        available = sorted(os.sched_getaffinity(0))
        if len(available) < 6:
            raise ValueError('six available logical CPUs required for separate 4/2 affinities')
        self.cpus = {'sut':available[:4], 'generator':available[4:6]}
        for name, memory, quota in [('sut',3*GIB,400000),('generator',GIB,200000)]:
            group = self.root/name
            group.mkdir()
            (group/'memory.max').write_text(str(memory))
            (group/'memory.swap.max').write_text('0')
            (group/'cpu.max').write_text(f'{quota} 100000')
            if int((group/'memory.max').read_text()) != memory:
                raise ValueError('memory control readback mismatch')

    def attach(self, name):
        def setup():
            (self.root/name/'cgroup.procs').write_text(str(os.getpid()))
            os.sched_setaffinity(0, self.cpus[name])
            resource.setrlimit(resource.RLIMIT_FSIZE,(FILE_LIMIT,FILE_LIMIT))
            resource.setrlimit(resource.RLIMIT_CORE,(0,0))
            os.umask(0o077)
        return setup

    def sample(self):
        return {name:dict(cpus=cpus, **{key:(self.root/name/key).read_text().strip()
                for key in ('memory.current','memory.peak','memory.max','memory.events','cpu.max','memory.swap.max')})
                for name, cpus in self.cpus.items()}


class Timex(ctypes.Structure):
    # Linux 64-bit struct timex, including ABI padding. Read-only adjtimex modes=0.
    _fields_ = [('modes',ctypes.c_uint), ('offset',ctypes.c_long), ('freq',ctypes.c_long),
                ('maxerror',ctypes.c_long), ('esterror',ctypes.c_long), ('status',ctypes.c_int),
                ('constant',ctypes.c_long), ('precision',ctypes.c_long), ('tolerance',ctypes.c_long),
                ('tv_sec',ctypes.c_long), ('tv_usec',ctypes.c_long), ('tick',ctypes.c_long),
                ('ppsfreq',ctypes.c_long), ('jitter',ctypes.c_long), ('shift',ctypes.c_int),
                ('stabil',ctypes.c_long), ('jitcnt',ctypes.c_long), ('calcnt',ctypes.c_long),
                ('errcnt',ctypes.c_long), ('stbcnt',ctypes.c_long), ('tai',ctypes.c_int),
                ('padding',ctypes.c_int*11)]


def clock_sample():
    if ctypes.sizeof(ctypes.c_long) != 8:
        raise ValueError('clock probe currently requires Linux 64-bit ABI')
    before = time.monotonic_ns()
    real = time.time_ns()
    after = time.monotonic_ns()
    tx = Timex()
    state = ctypes.CDLL(None,use_errno=True).adjtimex(ctypes.byref(tx))
    return dict(realtime_ns=real, monotonic_before_ns=before, monotonic_after_ns=after,
                kernel_clock_state=state, kernel_status=tx.status, maxerror_us=tx.maxerror,
                resolution_ns=math.ceil(time.get_clock_info('time').resolution*NS))


def clock_verdict(samples):
    # The realtime clock is shared by pcap and engine. This conservative bound
    # includes sampled realtime/monotonic drift and syscall brackets. NTP absolute
    # UTC error is common to both endpoints and is retained separately. The
    # interval includes pcap quantization; samples cannot exclude hidden paired steps.
    lows = [s['realtime_ns']-s['monotonic_after_ns'] for s in samples]
    highs = [s['realtime_ns']-s['monotonic_before_ns'] for s in samples]
    bound = max(highs)-min(lows)+1000+max(s['resolution_ns'] for s in samples)
    checked = all(s['kernel_clock_state'] >= 0 for s in samples) and bound <= 10_000_000
    return dict(clock_checked=checked, clock_uncertainty_ns=bound if checked else None,
                candidate_bound_ns=bound, method='shared host realtime interval; sampled monotonic brackets + resolution + pcap 1us; NTP absolute error retained separately',
                limitation='Sampling cannot exclude a step and inverse step entirely between probes; retain samples for review.')


def stopped(process, timeout=5):
    wake_error = None
    if process.poll() is None:
        os.killpg(process.pid, signal.SIGTERM)
        wake = getattr(process,'fixture_wake',None)
        if wake is not None:
            process.fixture_wake = None
            try:
                wake()
            except Exception as exc:
                wake_error = exc
    try:
        code = process.wait(timeout=timeout)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid,signal.SIGKILL)
        process.wait()
        raise RuntimeError('owned process exceeded graceful stop deadline; killed, evidence incomplete') from wake_error
    if wake_error is not None:
        raise RuntimeError('control wake failed; owned child still reaped') from wake_error
    return code


def wait_ready(processes, predicate, timeout=5):
    until = time.monotonic()+timeout
    while True:
        if any(p.poll() is not None for p in processes):
            raise RuntimeError('child exited before readiness')
        if predicate():
            return
        if time.monotonic() >= until:
            raise RuntimeError('readiness deadline expired')
        time.sleep(.025)


def health(port):
    with urllib.request.urlopen(f'http://127.0.0.1:{port}/api/health?verbose=true',timeout=.5) as response:
        return json.load(response)


def require_private_net(parent_inode):
    if parent_inode is None or os.stat('/proc/self/ns/net').st_ino == parent_inode:
        raise ValueError('refuse host network namespace')
    mapping = Path('/proc/self/uid_map').read_text().split()
    if len(mapping) != 3 or mapping[0] != '0' or mapping[2] != '1' or mapping[1] == '0':
        raise ValueError('requires single-user rootless namespace mapping')
    names = {name for _,name in socket.if_nameindex()}
    if names != {'lo'}:
        raise ValueError('namespace must start with loopback only')


def ethernet_ready():
    for args in [('link','add','nsgen','type','veth','peer','name','nssut'),
                 ('link','set','nsgen','address',LINK['src_mac']),
                 ('link','set','nssut','address',LINK['dst_mac']),
                 ('link','set','nsgen','up'),('link','set','nssut','up'),('link','set','lo','up')]:
        subprocess.run(['ip',*args],check=True,capture_output=True)
    # Verify the actual peer receives exact Ethernet bytes before starting SUT.
    with socket.socket(socket.AF_PACKET,socket.SOCK_RAW,socket.htons(0x88b5)) as incoming, \
         socket.socket(socket.AF_PACKET,socket.SOCK_RAW) as outgoing:
        incoming.bind(('nssut',0)); incoming.settimeout(1)
        outgoing.bind(('nsgen',0))
        probe = bytes.fromhex('02000000000202000000000188b5')+b'fixture-readiness'.ljust(46,b'.')
        if outgoing.send(probe) != len(probe):
            raise RuntimeError('short Ethernet readiness probe submission')
        observed, address = incoming.recvfrom(2048)
        if observed != probe or address[2] == socket.PACKET_OUTGOING:
            raise RuntimeError('exact inbound Ethernet peer proof failed')
    return dict(inbound=True, bytes=len(observed), frame_sha256=hashlib.sha256(observed).hexdigest(),
                namespace_inode=os.stat('/proc/self/ns/net').st_ino)


def send(run, run_id, smoke, origin_ns):
    # Separate generator process/cgroup. Intent is fsynced before each send batch
    # closes; abrupt death leaves unmatched intent, never presumed submitted.
    with socket.socket(socket.AF_PACKET,socket.SOCK_RAW) as sock, \
         (run/'intent.jsonl').open('xb',buffering=0) as intent, \
         (run/'offered.jsonl').open('xb',buffering=0) as offers, \
         (run/'submissions.jsonl').open('xb',buffering=0) as submissions:
        sock.bind(('nsgen',0))
        for seq, offset, positive in schedule(smoke):
            pid, raw, expected = packet(run_id,seq,positive)
            while (remaining := origin_ns+offset-time.time_ns()) > 0:
                time.sleep(min(remaining/NS,.02))
            at = time.time_ns()-origin_ns
            row = dict(packet_id=pid,offered_ns=at,offered_bytes=len(raw),expected_alerts=expected)
            intent.write(encode(dict(row,scheduled_ns=offset)))
            start = time.time_ns()-origin_ns
            error = None
            try:
                if sock.send(raw) != len(raw):
                    raise OSError('short frame submission')
            except OSError as exc:
                error = str(exc)
            end = time.time_ns()-origin_ns
            submissions.write(encode(dict(packet_id=pid,submitted=error is None,send_start_ns=start,
                send_return_ns=end,scheduled_ns=offset,lateness_ns=at-offset,error=error,
                frame_sha256=hashlib.sha256(raw).hexdigest())))
            offers.write(encode(dict(row,submitted=error is None)))
            if error:
                raise RuntimeError('submission failed; retained error and offered denominator')
        for stream in (intent,offers,submissions):
            os.fsync(stream.fileno())


def config(run, port, uds):
    # JSON quoted strings are valid YAML strings, including paths with spaces.
    fields = dict(uds_socket_path=str(uds),uds_socket_mode='0600',worker_count=1,
        channel_buffer_size=10000,db_dir=str(run),db_path=str(run/'alerts.db'),db_shard_daily=False,
        db_journal_mode='WAL',db_busy_timeout=5000,rules_seed_file=str(run/'rules.json'),
        suppressions_file=str(run/'suppressions.json'),api_listen_host='127.0.0.1',api_port=port,
        api_auth_enabled=False,pprof_enabled=False)
    return ('engine:\n'+''.join(f'  {k}: {json.dumps(v)}\n' for k,v in fields.items())+
            'logging:\n  format: "json"\n').encode()


def inspect_persistence(run):
    # Observe using independently encoded read-only URI after graceful stop.
    with sqlite3.connect((run/'alerts.db').as_uri()+'?mode=ro',uri=True) as db:
        integrity = db.execute('PRAGMA integrity_check').fetchone()[0]
        rows = db.execute('SELECT rule_id,sum(aggregated_count) FROM alerts GROUP BY rule_id ORDER BY rule_id').fetchall()
        unique_events = db.execute('SELECT count(*) FROM alert_events').fetchone()[0]
    return dict(integrity=integrity,counts_by_rule=rows,unique_events=unique_events)


def audit(run, run_id, smoke, persistence):
    """Compare exact independently regenerated identities, retaining every anomaly."""
    expected = {}
    offered = {}
    submissions = {}
    errors = []
    for filename, target in [('offered.jsonl',offered),('submissions.jsonl',submissions)]:
        with (run/filename).open() as stream:
            for line in stream:
                row = json.loads(line)
                key = row['packet_id']
                if key in target:
                    errors.append(f'duplicate {filename} {key}')
                target[key] = row
    counts = collections.Counter()
    for seq, offset, positive in schedule(smoke):
        pid, raw, events = packet(run_id,seq,positive)
        row = offered.get(pid)
        submission = submissions.get(pid)
        if row is None or submission is None or not row['submitted'] or not submission['submitted']:
            errors.append(f'missing/failed submission {pid}')
            continue
        if row['expected_alerts'] != events or row['offered_bytes'] != 512 or submission['scheduled_ns'] != offset:
            errors.append(f'changed oracle/schedule {pid}')
        for event in events:
            expected[event['event_id']] = event['rule_id']
            counts[event['rule_id']] += 1
    lifecycle = collections.Counter()
    durable = collections.Counter()
    arrivals = {}
    completions = {}
    with (run/'engine/events.jsonl').open() as stream:
        for line in stream:
            row = json.loads(line)
            if row['kind'] == 'durable':
                durable[row['event_id']] += 1
                completions[row['event_id']] = row['at_ns']
            else:
                lifecycle[(row['kind'],row['packet_id'])] += 1
                if row['kind'] == 'arrival':
                    arrivals[row['packet_id']] = row['at_ns']
    planned = sum(1 for _ in schedule(smoke))
    if len(offered) != planned or len(submissions) != planned:
        errors.append('offer/submission denominator differs')
    wanted = collections.Counter({(kind,pid):1 for pid in offered for kind in ('arrival','processed')})
    if lifecycle != wanted:
        errors.append('arrival/processed identity multiset differs')
    if durable != collections.Counter({key:1 for key in expected}):
        errors.append('durable identity multiset differs')
    if dict(persistence['counts_by_rule']) != dict(counts) or persistence['unique_events'] != sum(counts.values()):
        errors.append('independent read-only SQLite counts differ')
    capture = json.loads((run/'capture.json').read_text())
    if not capture['capture_closed'] or capture['sent'] != planned or any(capture[key] for key in
            ('dropped','parse_errors','invalid_markers','pcap_dropped','pcap_interface_dropped')):
        errors.append('capture incomplete or packet errors')
    api = json.loads((run/'api.json').read_text())
    throughput = api['throughput']
    if throughput['packets_received'] != planned or throughput['packets_processed'] != planned or throughput['alerts_generated'] != sum(counts.values()):
        errors.append('API packet/alert counts differ')
    delays = [completions[e['event_id']]-arrivals[pid] for pid,row in offered.items()
              for e in row['expected_alerts'] if e['event_id'] in completions and pid in arrivals]
    return dict(consistent=not errors,errors=errors,offered=len(offered),submitted=len(submissions),
                missing_durable=sorted(set(expected)-set(durable)),unexpected_durable=sorted(set(durable)-set(expected)),
                duplicate_durable=sorted(key for key,count in durable.items() if count>1),
                host_arrival_to_wal_return_ns=delays,
                oracle_events=len(expected),durable_events=sum(durable.values()),
                lifecycle_rows=sum(lifecycle.values()),sqlite_events=persistence['unique_events'])


def collection_manifest(run, result):
    provenance = json.loads((run/'provenance.json').read_text())
    declaration = json.loads((run/'frozen.json').read_text())
    capture = json.loads((run/'capture.json').read_text())
    return dict(schema_version=2,contract_id='local-v1',profile='local-v1',
        run_id=result['run_id'],started_at=result['origin'],duration_seconds=600,
        drain_seconds=30,observed_through_ns=630*NS,execution_complete=True,
        policy=dict(minimum_duration_seconds=600,minimum_expected_alerts=600),
        measurement=dict(live_arrival=True,durable_completion=True,
            clock_method=result['clock']['method'],clock_checked=result['clock']['clock_checked'],
            clock_uncertainty_ns=result['clock']['clock_uncertainty_ns']),
        resources=dict(sut_vcpus=4,sut_memory_bytes=3*GIB,active_rules=100,
                       generator_vcpus=2,generator_memory_bytes=GIB),
        provenance=dict(sut_commit=provenance['commit'],harness_commit=provenance['commit'],
            config_sha256=provenance['config_sha256'],rules_sha256=provenance['rules_sha256'],
            fixture_sha256=declaration['fixture_sha256']),
        integrity=dict(capture_errors=sum(capture[key] for key in ('dropped','parse_errors','invalid_markers')),
                       submission_errors=0,raw_evidence_complete=True,identity_accounting_complete=True),
        packet_windows=[dict(start_second=i*60,phase='burst' if i in (3,8) else 'sustained',
                             observation_complete=True) for i in range(10)])


def lock_root(root):
    descriptor = os.open(root,os.O_RDONLY|os.O_DIRECTORY|os.O_NOFOLLOW)
    try:
        fcntl.flock(descriptor,fcntl.LOCK_EX|fcntl.LOCK_NB)
    except BaseException:
        os.close(descriptor)
        raise
    return descriptor


def execute(root, run_id, smoke, parent_inode):
    require_private_net(parent_inode)
    # Inside rootless namespace os.getuid()=0; directory ownership is mapped to 0.
    previously_used = admit(root)
    run = root/run_id
    run.mkdir(mode=0o700)  # exclusive, no deletion on any error or success
    space = os.statvfs(root)
    meminfo = dict(line.split(':',1) for line in Path('/proc/meminfo').read_text().splitlines())
    save(run/'preflight.json',dict(kernel=os.uname().release,architecture=os.uname().machine,
        available_cpus=sorted(os.sched_getaffinity(0)),
        memory={key:meminfo[key].strip() for key in ('MemTotal','MemAvailable','SwapTotal')},
        disk_available_bytes=space.f_bavail*space.f_frsize,previous_evidence_bytes=previously_used,
        reservation_bytes=RUN_RESERVE,cumulative_limit_bytes=BUDGET))
    freeze_files(run,run_id,smoke)
    save(run/'suppressions.json',{'suppressions':[]})
    resource.setrlimit(resource.RLIMIT_FSIZE,(FILE_LIMIT,FILE_LIMIT))
    processes = []
    clocks = []
    result = dict(status='inconclusive',execution_complete=False,slo_compliance_asserted=False,
                  mode='smoke' if smoke else 'L1-fixture',error=None)
    groups = None
    # Short private UDS base avoids sockaddr_un length failures; keep its anchor
    # until cleanup and never unlink arbitrary replacement occupants.
    import tempfile
    socket_dir = Path(tempfile.mkdtemp(prefix='nsf-',dir='/tmp'))
    uds = socket_dir/'s'
    try:
        try:
            groups = Groups()
            result['ethernet'] = ethernet_ready()
            with socket.socket() as reservation:
                reservation.bind(('127.0.0.1',0)); port = reservation.getsockname()[1]
            origin_ns = (time.time_ns()//NS+6)*NS
            origin = datetime.fromtimestamp(origin_ns//NS,timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
            result.update(run_id=run_id,origin=origin)
            (run/'config.yaml').write_bytes(config(run,port,uds))
            save(run/'provenance.json',dict(source_sha256={str(p.relative_to(REPO)):digest(p) for p in
                [Path(__file__),REPO/'scripts/slo_ingress.py',REPO/'bin/netsentry-engine',REPO/'bin/netsentry-capture']},
                commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=REPO,text=True).strip(),
                config_sha256=digest(run/'config.yaml'),rules_sha256=digest(run/'rules.json')))
            def spawn(name, command, group):
                with (run/f'{name}.log').open('xb',buffering=0) as log:
                    process = subprocess.Popen(command,stdout=log,stderr=log,start_new_session=True,
                                               preexec_fn=groups.attach(group))
                processes.append(process)
                return process
            engine = spawn('engine',[str(REPO/'bin/netsentry-engine'),'-config',str(run/'config.yaml'),
                '--slo-output-dir',str(run/'engine'),'--slo-run-id',run_id,'--slo-origin',origin],'sut')
            def engine_ready():
                try:
                    return health(port)['engine']['rules_loaded'] == 100 and uds.is_socket()
                except (OSError,ValueError,KeyError):
                    return False
            wait_ready([engine],engine_ready)
            capture = spawn('capture',[str(REPO/'bin/netsentry-capture'),'-i','nssut','-s',str(uds),'-c','1',
                '--slo-run-id',run_id,'--slo-port',str(LINK['dst_port']),'--slo-output',str(run/'capture.json')],'sut')
            def wake_capture():
                # libpcap's read timeout need not wake a completely idle veth.
                # Separate control traffic carries a different run marker and
                # never enters the offered/eligible measurement denominator.
                raw = frame('fixture-control','wake',b'teardown',LINK,1)
                with socket.socket(socket.AF_PACKET,socket.SOCK_RAW) as sock:
                    sock.bind(('nsgen',0))
                    sent = sock.send(raw)
                save(run/'teardown.json',dict(run_id='fixture-control',bytes=sent,
                    frame_sha256=hashlib.sha256(raw).hexdigest(),purpose='wake idle pcap after SIGTERM'))
            capture.fixture_wake = wake_capture
            wait_ready([engine,capture],lambda:b'[capture] starting capture' in (run/'capture.log').read_bytes())
            if time.time_ns() >= origin_ns:
                raise RuntimeError('readiness missed frozen origin; no measurement offered')
            result['ready_before_origin'] = True
            clocks.append(clock_sample())
            if not clock_verdict(clocks)['clock_checked']:
                raise RuntimeError('clock probe unavailable before offering')
            generator = spawn('generator',[sys.executable,str(Path(__file__).resolve()),'_send',
                '--root',str(root),'--run-id',run_id,'--origin-ns',str(origin_ns)]+(['--smoke'] if smoke else []),'generator')
            finish = origin_ns+(5 if smoke else 630)*NS
            monotonic_stop = time.monotonic()+max(0,(finish-time.time_ns())/NS)+1
            with (run/'samples.jsonl').open('xb',buffering=0) as samples:
                while time.time_ns() < finish:
                    if time.monotonic() >= monotonic_stop:
                        raise RuntimeError('monotonic workload deadline exceeded')
                    if engine.poll() is not None or capture.poll() is not None:
                        raise RuntimeError('SUT exited during workload/observation')
                    if generator.poll() not in (None,0):
                        raise RuntimeError('generator failed; retain intent and submission ledgers')
                    used,_ = inventory(root); run_bytes,count = inventory(run)
                    if used > BUDGET or run_bytes > RUN_RESERVE or count > MAX_FILES:
                        raise RuntimeError('cumulative evidence budget/file envelope exceeded')
                    probe = clock_sample(); clocks.append(probe)
                    if not clock_verdict(clocks)['clock_checked']:
                        raise RuntimeError('clock interval bound invalidated during workload')
                    affinity = {str(p.pid):sorted(os.sched_getaffinity(p.pid)) for p in processes if p.poll() is None}
                    sample = dict(clock=probe,groups=groups.sample(),process_affinity=affinity,evidence_bytes=used,run_bytes=run_bytes)
                    samples.write(encode(sample))
                    time.sleep(.1)
                os.fsync(samples.fileno())
            if generator.wait(timeout=1) != 0:
                raise RuntimeError('generator did not complete')
            save(run/'api.json',health(port))
            result['returncodes'] = dict(capture=stopped(capture),engine=stopped(engine))
            if any(result['returncodes'].values()):
                raise RuntimeError('SUT did not close successfully')
            result['persistence'] = inspect_persistence(run)
            result['accounting'] = audit(run,run_id,smoke,result['persistence'])
            result['clock'] = clock_verdict(clocks)
            result['resources'] = groups.sample()
            if not result['clock']['clock_checked']:
                raise RuntimeError('shared-clock interval proof unavailable; retain raw samples')
            if smoke and not result['accounting']['consistent']:
                raise RuntimeError('smoke oracle/lifecycle/persistence accounting differs')
            if not smoke:
                save(run/'manifest.json',collection_manifest(run,result))
            result.update(status='smoke_complete' if smoke else 'collected_review_required',execution_complete=True)
        except Exception as exc:
            result['error'] = f'{type(exc).__name__}: {exc}'
        finally:
            for process in reversed(processes):
                try:
                    if process.poll() is None:
                        stopped(process)
                except Exception as exc:
                    result.update(status='inconclusive',execution_complete=False,error=str(exc))
            if clocks and 'clock' not in result:
                result['clock'] = clock_verdict(clocks)
            result['files'] = {str(p.relative_to(run)):dict(bytes=p.stat().st_size,sha256=digest(p))
                               for p in run.rglob('*') if p.is_file()}
            save(run/'result.json',result)
    finally:
        # Never recursively remove an occupant substituted for the owned socket.
        try:
            socket_dir.rmdir()
        except OSError:
            result['socket_cleanup'] = 'nonempty directory retained'
    return result


def derive(root, run_id):
    """Budgeted adapter copies and SQLite scratch stay under the same root lock."""
    if __package__:
        from . import slo_collect, slo_report
    else:
        import slo_collect, slo_report
    lock = lock_root(root)
    try:
        admit(root)
        run = root/run_id
        source = json.loads((run/'result.json').read_text())
        if not source['execution_complete'] or source['mode'] != 'L1-fixture':
            raise ValueError('derive requires a finalized full fixture, never smoke')
        for relative in ('manifest.json','offered.jsonl','engine/events.jsonl'):
            if digest(run/relative) != source['files'][relative]['sha256']:
                raise ValueError('sealed raw source changed')
        # Eight bounded adapter/scratch files at 64MiB fit the GiB reservation.
        resource.setrlimit(resource.RLIMIT_FSIZE,(64*1024**2,64*1024**2))
        scratch = run/'scratch'; scratch.mkdir(mode=0o700)
        slo_collect.collect(run/'manifest.json',run/'offered.jsonl',run/'engine/events.jsonl',
                            run/'adapter',scratch_dir=scratch,max_input_bytes=96*1024**2)
        document, _ = slo_report.read_observations(run/'adapter/observations.json')
        summary = slo_report.summarize(document)
        save(run/'adapter/report.json',summary)
        if inventory(root)[0] > BUDGET:
            raise RuntimeError('cumulative evidence budget exceeded')
        return summary
    finally:
        os.close(lock)


def launch(command, environment, unit):
    process = subprocess.Popen(command,env=environment,start_new_session=True)
    try:
        return process.wait()
    except (InterruptedError,KeyboardInterrupt):
        # Forward cancellation to the isolated supervisor, which finalizes its
        # separately grouped children. Never just kill the systemd-run client.
        if process.poll() is None:
            os.killpg(process.pid,signal.SIGTERM)
        try:
            process.wait(timeout=20)
        except subprocess.TimeoutExpired:
            subprocess.run(['systemctl','--user','kill','--signal=KILL',unit],
                           env=environment,check=False,capture_output=True)
            process.wait(timeout=5)
        raise


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('action',choices=['dry-run','smoke','run','derive','_run','_send'])
    parser.add_argument('--root',required=True,type=Path,help='existing private evidence root; all copies/scratch must live here')
    parser.add_argument('--run-id',required=True)
    parser.add_argument('--smoke',action='store_true',help=argparse.SUPPRESS)
    parser.add_argument('--origin-ns',type=int,help=argparse.SUPPRESS)
    parser.add_argument('--parent-net-inode',type=int,help=argparse.SUPPRESS)
    args = parser.parse_args(argv)
    if args.action != '_send':
        frozen(args.run_id,args.action == 'smoke' or args.smoke)
    root = args.root.expanduser().absolute()
    if args.action == 'derive':
        result = derive(root,args.run_id)
        print(json.dumps(result))
        return 0
    if args.action == '_send':
        mapping = Path('/proc/self/uid_map').read_text().split()
        if len(mapping) != 3 or mapping[0] != '0' or mapping[2] != '1' or mapping[1] == '0':
            raise ValueError('sender requires rootless mapping')
        if {name for _,name in socket.if_nameindex()} != {'lo','nsgen','nssut'}:
            raise ValueError('sender requires the owned namespace interfaces')
        send(root/args.run_id,args.run_id,args.smoke,args.origin_ns)
        return 0
    if args.action == '_run':
        lock = lock_root(root)
        try:
            result = execute(root,args.run_id,args.smoke,args.parent_net_inode)
        finally:
            os.close(lock)
        print(json.dumps({k:result[k] for k in ('status','execution_complete','slo_compliance_asserted','error')}))
        return 0 if result['execution_complete'] else 2
    admit(root)
    if args.action == 'dry-run':
        lock = lock_root(root)
        try:
            admit(root)
            run = root/args.run_id; run.mkdir(mode=0o700)
            freeze_files(run,args.run_id)
        finally:
            os.close(lock)
        print('Frozen no-send fixture; no SUT or socket opened; no L1/L2 acceptance.')
        return 0
    for name in ('netsentry-engine','netsentry-capture'):
        if not (REPO/'bin'/name).is_file():
            raise ValueError('run make build before live fixture')
    environment = dict(os.environ)
    environment.setdefault('XDG_RUNTIME_DIR',f'/run/user/{os.getuid()}')
    unit = 'run-nsf-'+uuid.uuid4().hex+'.scope'
    command = ['systemd-run','--user','--scope','--quiet','--unit='+unit,'-p','Delegate=yes',
               'unshare','--user','--map-root-user','--net','--',sys.executable,str(Path(__file__).resolve()),
               '_run','--root',str(root),'--run-id',args.run_id,'--parent-net-inode',str(os.stat('/proc/self/ns/net').st_ino)]
    if args.action == 'smoke':
        command += ['--smoke']
    return launch(command,environment,unit)


if __name__ == '__main__':
    def cancel(signum, _frame):
        signal.signal(signal.SIGTERM,signal.SIG_IGN)
        signal.signal(signal.SIGINT,signal.SIG_IGN)
        raise InterruptedError(f'cancelled by signal {signum}; partial evidence retained')
    signal.signal(signal.SIGTERM,cancel)
    signal.signal(signal.SIGINT,cancel)
    try:
        raise SystemExit(main())
    except (OSError,ValueError,RuntimeError) as error:
        print(f'local-fixture: {error}',file=sys.stderr)
        raise SystemExit(2)
