"""Direct fixture regressions; synthetic tests are never measured L1/L2."""
from __future__ import annotations

import collections
import hashlib
import json
import os
from pathlib import Path
import socket
import signal
import sqlite3
import struct
import subprocess
import sys
import tempfile
import time
import unittest
from unittest import mock

from . import local_fixture as fixture


_REAL_DIGEST = fixture.digest


def python_child_digest(path):
    # These lifecycle tests launch Python sleepers instead of either SUT binary.
    # Bind their provenance to that real executable; a clean checkout has no bin.
    if path in (fixture.REPO/'bin/netsentry-engine',fixture.REPO/'bin/netsentry-capture'):
        return _REAL_DIGEST(Path(sys.executable))
    return _REAL_DIGEST(path)


class FixtureTests(unittest.TestCase):
    def test_frozen_schedule_and_independent_packet_oracle(self):
        rows = list(fixture.schedule())
        self.assertEqual(len(rows),108000)
        minute_counts = collections.Counter(offset//(60*fixture.NS) for _,offset,_ in rows)
        self.assertEqual([minute_counts[i] for i in range(10)],[6000,6000,6000,30000,6000,6000,6000,6000,30000,6000])
        positives = [(offset//fixture.NS,positive) for _,offset,positive in rows if positive is not None]
        self.assertEqual(positives,[(i,i%100) for i in range(600)])
        rules = fixture.rules()['rules']
        self.assertEqual(len(rules),100)
        self.assertEqual(len({rule['id'] for rule in rules}),100)
        keywords = [rule['config']['keywords'][0].encode() for rule in rules]
        def checksum(raw):
            if len(raw)%2: raw += b'\0'
            total = sum(struct.unpack('!%dH'%(len(raw)//2),raw))
            while total>>16: total=(total&65535)+(total>>16)
            return total
        # Reach every distinct positive and ordinary negative frame plus IP-ID wrap.
        for index in list(range(101)) + [65537]:
            positive = index if index<100 else None
            pid,raw,events = fixture.packet('independent',index+1,positive)
            self.assertEqual(len(raw),512)
            self.assertEqual(raw[12:14],b'\x08\x00')
            self.assertEqual(raw[14],0x45)
            self.assertEqual(raw[20:22],b'\x40\x00')
            self.assertEqual(raw[23],17)
            self.assertEqual(checksum(raw[14:34]),65535)
            length = struct.unpack('!H',raw[38:40])[0]
            pseudo = raw[26:34]+b'\0\x11'+struct.pack('!H',length)
            self.assertEqual(checksum(pseudo+raw[34:]),65535)
            matches = [i for i,keyword in enumerate(keywords) if keyword in raw[42:]]
            self.assertEqual(matches,[] if positive is None else [positive])
            self.assertIn(f'NSLO1 independent {pid}\n'.encode(),raw[42:])
            if positive is not None:
                rule = f'local-{positive:03}'
                self.assertEqual(events,[dict(rule_id=rule,event_id='slo_'+hashlib.sha256((pid+'\0'+rule).encode()).hexdigest())])
        first=fixture.frozen('frozen');second=fixture.frozen('frozen')
        self.assertEqual(first,second)
        self.assertEqual(first['planned_packets'],108000)
        self.assertNotEqual(first['fixture_sha256'],fixture.frozen('other')['fixture_sha256'])

    def test_dry_run_never_opens_socket_or_starts_process_and_preserves_existing(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            with mock.patch.object(fixture.socket,'socket',side_effect=AssertionError('network opened')), \
                 mock.patch.object(fixture.subprocess,'Popen',side_effect=AssertionError('process started')):
                self.assertEqual(fixture.main(['dry-run','--root',directory,'--run-id','dry']),0)
                before={p.name:p.read_bytes() for p in (root/'dry').iterdir()}
                with self.assertRaises(FileExistsError):
                    fixture.main(['dry-run','--root',directory,'--run-id','dry'])
                self.assertEqual(before,{p.name:p.read_bytes() for p in (root/'dry').iterdir()})
            self.assertFalse(json.loads((root/'dry/frozen.json').read_text())['execution_complete'])

    def test_budget_disk_links_and_serialized_root_ownership(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            with mock.patch.object(fixture,'inventory',return_value=(3*fixture.GIB,1)):
                self.assertEqual(fixture.admit(root,8*fixture.GIB),3*fixture.GIB)
            with mock.patch.object(fixture,'inventory',return_value=(3*fixture.GIB+1,1)):
                with self.assertRaisesRegex(ValueError,'4GiB'):
                    fixture.admit(root,8*fixture.GIB)
            with self.assertRaisesRegex(ValueError,'8GiB'):
                fixture.admit(root,8*fixture.GIB-1)
            (root/'neighbor').write_text('preserve')
            (root/'alias').symlink_to(root/'neighbor')
            with self.assertRaisesRegex(ValueError,'symlink'):
                fixture.admit(root,8*fixture.GIB)
            (root/'alias').unlink()
            os.link(root/'neighbor',root/'alias')
            with self.assertRaisesRegex(ValueError,'hard link'):
                fixture.admit(root,8*fixture.GIB)
            (root/'alias').unlink()
            lock=fixture.lock_root(root)
            try:
                with self.assertRaises(BlockingIOError): fixture.lock_root(root)
            finally: os.close(lock)
            self.assertEqual((root/'neighbor').read_text(),'preserve')
            with mock.patch.object(fixture.os,'getuid',return_value=root.stat().st_uid+1):
                with self.assertRaisesRegex(ValueError,'owned directory'): fixture.admit(root,8*fixture.GIB)
            os.mkfifo(root/'fifo')
            with self.assertRaisesRegex(ValueError,'unsupported evidence object'): fixture.admit(root,8*fixture.GIB)
            (root/'fifo').unlink()
            root.chmod(0o755)
            with self.assertRaisesRegex(ValueError,'0700'): fixture.admit(root,8*fixture.GIB)

    def test_child_setup_enforces_file_cap_and_actual_affinity(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);(root/'sut').mkdir();(root/'sut/cgroup.procs').touch()
            group=object.__new__(fixture.Groups);group.root=root
            group.cpus={'sut':[min(os.sched_getaffinity(0))]}
            code="import os,resource,json,sys; print(json.dumps(dict(cpus=sorted(os.sched_getaffinity(0)),limit=resource.getrlimit(resource.RLIMIT_FSIZE))),flush=True); open(sys.argv[1],'wb',buffering=0).write(b'x'*2048)"
            # A write may short-write at the hard cap, so check the resulting size
            # and actual readback rather than assuming an exception is immediate.
            with mock.patch.object(fixture,'FILE_LIMIT',1024):
                result=subprocess.run([sys.executable,'-c',code,str(root/'bounded')],
                    preexec_fn=group.attach('sut'),capture_output=True,text=True,check=False)
            readback=json.loads(result.stdout)
            self.assertEqual(readback['limit'],[1024,1024])
            self.assertEqual(readback['cpus'],group.cpus['sut'])
            self.assertLessEqual((root/'bounded').stat().st_size,1024)
            self.assertTrue((root/'sut/cgroup.procs').read_text().isdigit())

    def test_clock_shared_interval_bound_and_steps(self):
        sample=dict(realtime_ns=20000000,monotonic_before_ns=10000000,monotonic_after_ns=10001000,
                    kernel_clock_state=5,kernel_status=0x40,maxerror_us=16000000,resolution_ns=1)
        # Unsynchronized absolute UTC does not add a 16-second error to intervals
        # taken by two users of the same realtime clock. Keep raw NTP state.
        result=fixture.clock_verdict([sample,dict(sample,realtime_ns=21000000,monotonic_before_ns=11000000,monotonic_after_ns=11001000)])
        self.assertTrue(result['clock_checked'])
        self.assertLess(result['clock_uncertainty_ns'],10000)
        self.assertFalse(fixture.clock_verdict([sample,dict(sample,realtime_ns=40000000)])['clock_checked'])
        self.assertFalse(fixture.clock_verdict([dict(sample,kernel_clock_state=-1)])['clock_checked'])

    def test_submission_failure_retains_intent_offer_and_outcome(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            fake=mock.MagicMock()
            fake.__enter__.return_value.send.side_effect=OSError('injected send failure')
            with mock.patch.object(fixture.socket,'socket',return_value=fake), \
                 mock.patch.object(fixture,'schedule',return_value=iter([(1,0,0)])):
                with self.assertRaisesRegex(RuntimeError,'submission failed'):
                    fixture.send(root,'failure',True,time.time_ns()-fixture.NS)
            offer=json.loads((root/'offered.jsonl').read_text())
            intent=json.loads((root/'intent.jsonl').read_text())
            submitted=json.loads((root/'submissions.jsonl').read_text())
            self.assertFalse(offer['submitted'])
            self.assertEqual(intent['packet_id'],offer['packet_id'])
            self.assertEqual(len(offer['expected_alerts']),1)
            self.assertIn('injected send failure',submitted['error'])

    def test_readiness_early_failure_timeout_and_stop(self):
        process=subprocess.Popen([sys.executable,'-c','raise SystemExit(7)'],start_new_session=True)
        process.wait()
        with self.assertRaisesRegex(RuntimeError,'exited before readiness'):
            fixture.wait_ready([process],lambda:True)
        process=subprocess.Popen([sys.executable,'-c','import time;time.sleep(30)'],start_new_session=True)
        try:
            with self.assertRaisesRegex(RuntimeError,'deadline'):
                fixture.wait_ready([process],lambda:False,timeout=.05)
            fixture.wait_ready([process],lambda:True)
        finally:
            fixture.stopped(process)
        self.assertIsNotNone(process.returncode)

    def test_stop_deadline_still_reaps_owned_child(self):
        with tempfile.TemporaryDirectory() as directory:
            ready=Path(directory)/'ready'
            code="import signal,time,pathlib,sys; signal.signal(signal.SIGTERM,signal.SIG_IGN); pathlib.Path(sys.argv[1]).write_text('ready'); time.sleep(30)"
            child=subprocess.Popen([sys.executable,'-c',code,str(ready)],start_new_session=True)
            try:
                fixture.wait_ready([child],ready.exists)
                with self.assertRaisesRegex(RuntimeError,'killed'):
                    fixture.stopped(child,timeout=.05)
                self.assertIsNotNone(child.poll())
            finally:
                if child.poll() is None:
                    os.killpg(child.pid,signal.SIGKILL);child.wait()

    def test_cancel_before_readiness_preserves_neighbor_and_replacement(self):
        # Exercise execute's actual startup/finally with a live owned child.
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as sockets:
            root=Path(directory)
            neighbor=root/'neighbor';neighbor.write_text('untouched')
            replacement=Path(sockets)/'s';replacement.write_text('keep replacement')
            children=[]
            real_popen=subprocess.Popen
            def spawn(*args,**kwargs):
                kwargs.pop('preexec_fn')
                child=real_popen([sys.executable,'-c','import time;time.sleep(30)'],**kwargs)
                children.append(child)
                return child
            with mock.patch.object(fixture,'require_private_net'), \
                 mock.patch.object(fixture,'Groups'),mock.patch.object(fixture,'ethernet_ready',return_value={'inbound':True}), \
                 mock.patch.object(fixture.subprocess,'Popen',side_effect=spawn), \
                 mock.patch.object(fixture,'digest',side_effect=python_child_digest), \
                 mock.patch.object(fixture.subprocess,'check_output',return_value='a'*40), \
                 mock.patch.object(fixture,'wait_ready',side_effect=InterruptedError('cancel before readiness')), \
                 mock.patch('tempfile.mkdtemp',return_value=sockets), \
                 mock.patch.object(fixture.resource,'setrlimit'):
                result=fixture.execute(root,'cancel',True,1)
            self.assertFalse(result['execution_complete'])
            self.assertIn('cancel before readiness',result['error'])
            self.assertTrue(all(child.poll() is not None for child in children))
            self.assertEqual(neighbor.read_text(),'untouched')
            self.assertEqual(replacement.read_text(),'keep replacement')
            self.assertTrue((root/'cancel/result.json').is_file())

    def test_active_cancel_reaps_all_children_and_preserves_symlink_target(self):
        with tempfile.TemporaryDirectory() as directory, tempfile.TemporaryDirectory() as sockets:
            root=Path(directory)
            target=root/'target';target.write_text('preserve target')
            replacement=Path(sockets)/'s';replacement.symlink_to(target)
            children=[]
            real_popen=subprocess.Popen
            def spawn(*args,**kwargs):
                kwargs.pop('preexec_fn')
                child=real_popen([sys.executable,'-c','import time;time.sleep(30)'],**kwargs)
                children.append(child)
                return child
            fake_socket=mock.MagicMock()
            fake_socket.__enter__.return_value.getsockname.return_value=('127.0.0.1',12345)
            fake_socket.__enter__.return_value.send.return_value=60
            with mock.patch.object(fixture,'require_private_net'), \
                 mock.patch.object(fixture,'Groups'),mock.patch.object(fixture,'ethernet_ready',return_value={'inbound':True}), \
                 mock.patch.object(fixture.subprocess,'Popen',side_effect=spawn), \
                 mock.patch.object(fixture,'digest',side_effect=python_child_digest), \
                 mock.patch.object(fixture.subprocess,'check_output',return_value='a'*40), \
                 mock.patch.object(fixture,'wait_ready'), \
                 mock.patch.object(fixture.socket,'socket',return_value=fake_socket), \
                 mock.patch.object(fixture,'clock_sample',side_effect=[dict(realtime_ns=10000,monotonic_before_ns=1000,monotonic_after_ns=1001,kernel_clock_state=0,kernel_status=0,maxerror_us=0,resolution_ns=1),InterruptedError('active cancellation')]), \
                 mock.patch('tempfile.mkdtemp',return_value=sockets), \
                 mock.patch.object(fixture.resource,'setrlimit'):
                result=fixture.execute(root,'active',True,1)
            self.assertEqual(len(children),3)
            self.assertTrue(all(child.poll() is not None for child in children))
            self.assertFalse(result['execution_complete'])
            self.assertIn('active cancellation',result['error'])
            self.assertTrue(replacement.is_symlink())
            self.assertEqual(target.read_text(),'preserve target')

    def test_accounting_retains_missing_extra_and_duplicate_identities(self):
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory);(root/'engine').mkdir()
            offers=[];submissions=[];events=[]
            for sequence,offset,positive in fixture.schedule(True):
                pid,raw,expected=fixture.packet('audit',sequence,positive)
                offers.append(dict(packet_id=pid,offered_ns=offset,offered_bytes=len(raw),expected_alerts=expected,submitted=True))
                submissions.append(dict(packet_id=pid,submitted=True,scheduled_ns=offset))
                events.extend([dict(kind='arrival',packet_id=pid,at_ns=offset+1),dict(kind='processed',packet_id=pid,at_ns=offset+2)])
                events.extend(dict(kind='durable',event_id=e['event_id'],at_ns=offset+3) for e in expected)
            def ledger(name,rows):
                (root/name).write_bytes(b''.join(fixture.encode(row) for row in rows))
            ledger('offered.jsonl',offers);ledger('submissions.jsonl',submissions);ledger('engine/events.jsonl',events)
            (root/'capture.json').write_bytes(fixture.encode(dict(capture_closed=True,sent=20,dropped=0,parse_errors=0,invalid_markers=0,pcap_dropped=0,pcap_interface_dropped=0)))
            (root/'api.json').write_bytes(fixture.encode(dict(throughput=dict(packets_received=20,packets_processed=20,alerts_generated=2))))
            persistence=dict(counts_by_rule=[('local-000',1),('local-001',1)],unique_events=2)
            self.assertTrue(fixture.audit(root,'audit',True,persistence)['consistent'])
            durable=[row for row in events if row['kind']=='durable']
            ledger('engine/events.jsonl',[row for row in events if row!=durable[1]]+[durable[0],dict(kind='durable',event_id='unexpected',at_ns=100)])
            result=fixture.audit(root,'audit',True,persistence)
            self.assertFalse(result['consistent'])
            self.assertEqual(result['missing_durable'],[durable[1]['event_id']])
            self.assertEqual(result['duplicate_durable'],[durable[0]['event_id']])
            self.assertEqual(result['unexpected_durable'],['unexpected'])

    def test_manifest_matches_local_collector_contract(self):
        from . import slo_collect
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            (root/'provenance.json').write_bytes(fixture.encode(dict(commit='1'*40,config_sha256='2'*64,rules_sha256='3'*64)))
            (root/'frozen.json').write_bytes(fixture.encode(dict(fixture_sha256='4'*64)))
            (root/'capture.json').write_bytes(fixture.encode(dict(dropped=0,parse_errors=0,invalid_markers=0)))
            result=dict(run_id='manifest',origin='2026-10-18T00:00:00Z',clock=dict(method='synthetic',clock_checked=True,clock_uncertainty_ns=1000))
            manifest=fixture.collection_manifest(root,result)
            decoded=slo_collect._manifest(manifest)
            self.assertEqual(decoded['contract_id'],'local-v1')
            self.assertEqual(len(decoded['packet_windows']),10)

    def test_budgeted_adapter_replays_supplied_raw_and_rejects_drift(self):
        from .test_slo_local import sources
        with tempfile.TemporaryDirectory() as directory:
            root=Path(directory)
            sources(root/'sample',run_id='sample')
            run=root/'sample'
            (run/'engine').mkdir()
            (run/'events.jsonl').rename(run/'engine/events.jsonl')
            sealed={name:dict(sha256=fixture.digest(run/name)) for name in ('manifest.json','offered.jsonl','engine/events.jsonl')}
            fixture.save(run/'result.json',dict(execution_complete=True,mode='L1-fixture',files=sealed))
            with mock.patch.object(fixture.resource,'setrlimit'):
                summary=fixture.derive(root,'sample')
            self.assertFalse(summary['slo_compliance_asserted'])
            self.assertTrue((run/'adapter/receipt.json').is_file())
            self.assertEqual(list((run/'scratch').iterdir()),[])
            self.assertEqual((run/'offered.jsonl').read_bytes(),(run/'adapter/offered.jsonl').read_bytes())
            with (run/'offered.jsonl').open('ab') as stream: stream.write(b'\n')
            with self.assertRaisesRegex(ValueError,'sealed raw source changed'):
                fixture.derive(root,'sample')

    def test_launcher_forwards_cancel_to_owned_session(self):
        with tempfile.TemporaryDirectory() as directory:
            ready=Path(directory)/'ready';closed=Path(directory)/'closed'
            script="import signal,time,pathlib,sys; signal.signal(signal.SIGTERM,lambda *_:(pathlib.Path(sys.argv[2]).write_text('closed'),sys.exit(0))); pathlib.Path(sys.argv[1]).write_text('ready'); time.sleep(30)"
            child=subprocess.Popen([sys.executable,'-c',script,str(ready),str(closed)],start_new_session=True)
            try:
                fixture.wait_ready([child],ready.exists)
                real_wait=child.wait
                calls=0
                def wait(*args,**kwargs):
                    nonlocal calls
                    calls+=1
                    if calls==1: raise InterruptedError('outer cancel')
                    return real_wait(*args,**kwargs)
                with mock.patch.object(fixture.subprocess,'Popen',return_value=child), \
                     mock.patch.object(child,'wait',side_effect=wait):
                    with self.assertRaisesRegex(InterruptedError,'outer cancel'):
                        fixture.launch(['unused'],dict(os.environ),'run-owned.scope')
                self.assertEqual(closed.read_text(),'closed')
                self.assertEqual(child.poll(),0)
            finally:
                if child.poll() is None: fixture.stopped(child)

    def test_namespace_refuses_parent_or_unowned_map(self):
        inode=os.stat('/proc/self/ns/net').st_ino
        with self.assertRaisesRegex(ValueError,'host network'):
            fixture.require_private_net(inode)
        with mock.patch.object(Path,'read_text',return_value='0 0 4294967295'):
            with self.assertRaisesRegex(ValueError,'rootless'): fixture.require_private_net(inode+1)

    def test_readonly_persistence_with_encoded_path(self):
        with tempfile.TemporaryDirectory(prefix='fixture space ') as directory:
            root=Path(directory)
            with sqlite3.connect(root/'alerts.db') as db:
                db.executescript('CREATE TABLE alerts(rule_id TEXT,aggregated_count INTEGER); CREATE TABLE alert_events(event_id TEXT); INSERT INTO alerts VALUES ("local-001",2); INSERT INTO alert_events VALUES ("a"),("b");')
            before=(root/'alerts.db').read_bytes()
            result=fixture.inspect_persistence(root)
            self.assertEqual(result['counts_by_rule'],[('local-001',2)])
            self.assertEqual(result['unique_events'],2)
            self.assertEqual(result['integrity'],'ok')
            self.assertEqual(before,(root/'alerts.db').read_bytes())


if __name__=='__main__': unittest.main()
