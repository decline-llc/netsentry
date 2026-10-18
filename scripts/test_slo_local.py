"""Synthetic local-v1 evaluator/retention regressions; these are not L1/L2 measurements."""
from __future__ import annotations

import contextlib
import copy
import hashlib
import io
import json
from pathlib import Path
import subprocess
import sqlite3
import sys
import tempfile
import unittest
from unittest import mock

from . import slo_bundle as bundle, slo_collect as collect, slo_compare as compare
from . import slo_context as context, slo_context_compare as context_compare
from . import slo_local as local, slo_reconstruct as reconstruct, slo_report as report

NS = report.NS
FIXTURE = Path(__file__).with_name('fixtures') / 'slo_legacy_golden.json'


def encoded(value):
    return (json.dumps(value, indent=2) + '\n').encode()


def write(path, value):
    path.write_bytes(encoded(value))


def observation():
    doc = dict(schema_version=2, contract_id='local-v1', profile='local-v1', run_id='synthetic-run',
               started_at='2026-10-18T00:00:00Z', duration_seconds=600, drain_seconds=30,
               observed_through_ns=630*NS, execution_complete=True, policy=dict(local.POLICY),
               measurement=dict(live_arrival=True, durable_completion=True, clock_method='synthetic',
                                clock_uncertainty_ns=10_000_000, clock_checked=True),
               resources=dict(sut_vcpus=4, sut_memory_bytes=3*1024**3, active_rules=100,
                              generator_vcpus=2, generator_memory_bytes=1024**3),
               provenance=dict(sut_commit='1'*40, harness_commit='2'*40, config_sha256='3'*64,
                               rules_sha256='4'*64, fixture_sha256='5'*64),
               integrity=dict(capture_errors=0, submission_errors=0, raw_evidence_complete=True,
                              identity_accounting_complete=True), anomalies=[], packet_windows=[], expected_alerts=[])
    for minute in range(10):
        n = 30_000 if minute in (3, 8) else 6000
        doc['packet_windows'].append(dict(start_second=minute*60, phase='burst' if minute in (3, 8) else 'sustained',
                                          offered_eligible=n, submitted_eligible=n, fully_processed=n,
                                          offered_bytes=n*512, observation_complete=True))
    for second in range(600):
        doc['expected_alerts'].append(dict(event_id=f'e{second}', packet_id=f'p{second}', rule_id=f'r{second%100}',
                                          offered_ns=second*NS, arrival_ns=second*NS+1000,
                                          durable_ns=second*NS+490_001_000))
    return doc


def review(doc, raw):
    return dict(schema_version=2, contract_id='local-v1', run_id=doc['run_id'],
                observations_sha256=hashlib.sha256(raw).hexdigest(), reviewer='synthetic-reviewer',
                reviewed_at='2026-10-18T01:00:00Z',
                checks={key: True for key in local.REVIEW_CHECKS},
                evidence_sha256={key: 'a'*64 for key in ('bundle','context','raw_manifest')})


def manifest(doc):
    result = copy.deepcopy(doc)
    result.pop('expected_alerts')
    result.pop('anomalies', None)
    for row in result['packet_windows']:
        for key in ('offered_eligible','fully_processed','offered_bytes','submitted_eligible'):
            row.pop(key, None)
    return result


def sources(root, *, profile='local-v1', run_id='synthetic-run', start='2026-10-18T00:00:00Z', extra=()):
    """Deliberately sparse ledgers: test derivation, never valid offered-load evidence."""
    root.mkdir()
    doc = observation()
    doc.update(run_id=run_id, started_at=start)
    if profile != 'local-v1':
        legacy = json.loads(FIXTURE.read_text())['cases'][0]['input']
        doc.update(schema_version=1, profile=profile, policy=legacy['policy'], measurement=legacy['measurement'],
                   resources=legacy['resources'])
        for key in ('contract_id','anomalies','integrity'):
            del doc[key]
    offers, events = [], []
    for second in range(0,600,60):
        a = observation()['expected_alerts'][second]
        offer = dict(packet_id=a['packet_id'], offered_ns=a['offered_ns'], offered_bytes=512,
                     expected_alerts=[dict(event_id=a['event_id'], rule_id=a['rule_id'])])
        if profile == 'local-v1':
            offer['submitted'] = True
        offers.append(offer)
        events.extend([dict(kind='arrival',packet_id=a['packet_id'],at_ns=a['arrival_ns']),
                       dict(kind='durable',event_id=a['event_id'],at_ns=a['durable_ns']),
                       dict(kind='processed',packet_id=a['packet_id'],at_ns=a['durable_ns']+1000)])
    write(root/'manifest.json', manifest(doc))
    for name, rows in (('offered.jsonl',offers),('events.jsonl',events+list(extra))):
        (root/name).write_bytes(b''.join(json.dumps(row).encode()+b'\n' for row in rows))
    return doc


def adapt(root):
    return collect.collect(root/'manifest.json',root/'offered.jsonl',root/'events.jsonl',root/'adapter')


def retained_context(root, observations):
    doc, raw = report.read_observations(observations)
    proof = b'explicitly synthetic opaque context proof\n'
    evidence = root/'evidence'; evidence.mkdir()
    (evidence/'proof.bin').write_bytes(proof)
    declaration = {key: doc[key] for key in ('schema_version','profile','run_id','started_at')}
    if doc['schema_version'] == 2:
        declaration['contract_id'] = doc['contract_id']
    declaration.update(artifact_kind='netsentry_slo_run_context', observations_sha256=hashlib.sha256(raw).hexdigest(),
                       evidence=[dict(id='proof',description='synthetic test only',bytes=len(proof),sha256=hashlib.sha256(proof).hexdigest())])
    for section, fields in context.fields_for(doc).items():
        declaration[section] = {}
        for key, kind in fields.items():
            value = 'synthetic' if kind == 'text' else 1 if kind == 'positive_integer' else True if kind == 'boolean' else kind[0]
            if key in ('sut_vcpus','sut_memory_bytes'):
                value = doc['resources'][key]
            if key == 'deployment_type' and doc['schema_version'] == 2:
                value = 'vm'
            declaration[section][key] = dict(value=value,evidence_ids=['proof'])
    write(root/'declaration.json',declaration)
    return context.retain(root/'retained',declaration=root/'declaration.json',observations=observations,evidence_dir=evidence)


def package(root, *, profile='local-v1', run_id='synthetic-run', start='2026-10-18T00:00:00Z'):
    sources(root, profile=profile,run_id=run_id,start=start)
    # Syntactic sender envelope only; adapter replay is tested separately. No
    # claim of frame/oracle replay or authentic acquisition is made.
    fixture = b'{}\n'*10
    m = json.loads((root/'manifest.json').read_text())
    m['provenance']['fixture_sha256'] = hashlib.sha256(fixture).hexdigest()
    write(root/'manifest.json',m)
    adapt(root)
    doc, raw = report.read_observations(root/'adapter/observations.json')
    summary = report.summarize(doc)
    summary['source'] = dict(sha256=hashlib.sha256(raw).hexdigest(),raw_json=raw.decode())
    write(root/'report.json',summary)
    sender=root/'sender';sender.mkdir()
    (sender/'fixture.jsonl').write_bytes(fixture)
    (sender/'offered.jsonl').write_bytes((root/'offered.jsonl').read_bytes())
    (sender/'submissions.jsonl').write_bytes(b''.join(json.dumps(dict(packet_id=f'p{second}',submitted=True,
        send_start_ns=second*NS,send_return_ns=second*NS+1,scheduled_ns=second*NS,lateness_ns=0,
        frame_sha256='a'*64)).encode()+b'\n' for second in range(0,600,60)))
    entries=[]
    for name in bundle.SOURCES['sender'][1:]:
        data=(sender/name).read_bytes()
        entries.append(dict(file=name,bytes=len(data),rows=10,sha256=hashlib.sha256(data).hexdigest()))
    write(sender/'submission.json',dict(schema_version=1,run_id=run_id,origin=start,submission_complete=True,
        execution_complete=False,slo_compliance_asserted=False,departmental_review_required=True,
        link=dict(src_mac='02:00:00:00:00:01',dst_mac='02:00:00:00:00:02',src_ip='192.0.2.1',dst_ip='192.0.2.2',src_port=10000,dst_port=10001),
        eligible_byte_boundary='ethernet_frame_without_fcs_preamble_ifg_with_padding',inputs_and_outputs=entries,
        sender_source_sha256='a'*64,limitations=['synthetic envelope; no sender replay']))
    capture=dict(schema_version=1,run_id=run_id,capture_closed=True,execution_complete=False,
                 slo_compliance_asserted=False,timestamp_type='host',timestamp_precision='microseconds',direction='inbound',udp_port=10001,pcap_stats_valid=True)
    capture.update({key:0 for key in ('sent','dropped','parse_errors','ignored','invalid_markers','pcap_received','pcap_dropped','pcap_interface_dropped')})
    write(root/'capture.json',capture)
    engine=root/'engine';engine.mkdir()
    (engine/'events.jsonl').write_bytes((root/'events.jsonl').read_bytes())
    data=(engine/'events.jsonl').read_bytes()
    write(engine/'close.json',dict(schema_version=1,run_id=run_id,origin=start,export_closed=True,execution_complete=False,
        slo_compliance_asserted=False,departmental_review_required=True,file='events.jsonl',rows=30,bytes=len(data),
        sha256=hashlib.sha256(data).hexdigest(),clock='supplied_live_arrival_and_engine_unix_wall_clock',
        durable_boundary='successful_full_synchronous_WAL_WriteBatch_return_upper_bound',
        event_identity='slo_ plus SHA256(UTF8(packet_id) + NUL + UTF8(rule_id))'))
    result=bundle.reconcile(root/'bundle',sender=sender,capture=root/'capture.json',engine=engine,adapter=root/'adapter',summary=root/'report.json',reconstruct_ledgers=True)
    return result


class LocalEvaluationTests(unittest.TestCase):
    def test_frozen_legacy_outputs_and_diagnostics(self):
        for case in json.loads(FIXTURE.read_text())['cases']:
            with self.subTest(case=case['label']):
                self.assertEqual(report.summarize(case['input']),case['expected'])
        base=json.loads(FIXTURE.read_text())['cases'][0]['input']
        for key,value,message in [('profile','local-v1','profile must be staging or prod'),
                                  ('schema_version',2,'schema 2 requires profile and contract_id local-v1'),
                                  ('execution_complete',False,'completed-run report requires execution_complete=true')]:
            d=copy.deepcopy(base);d[key]=value
            with self.assertRaisesRegex(report.EvidenceError,message):report.summarize(d)
        with self.assertRaisesRegex(report.EvidenceError,'root: fields do not match schema v1'):
            report.summarize(dict(base,contract_id='local-v1'))

    def test_contract_identity_and_exact_members(self):
        for field,value in [('contract_id','local-v2'),('profile','staging'),('schema_version',1),('schema_version',True),('schema_version',3)]:
            d=observation();d[field]=value
            with self.subTest(field=field,value=value),self.assertRaises(report.EvidenceError):report.summarize(d)
        for section,key in [(None,'contract_id'),('measurement','clock_checked'),('integrity','capture_errors'),('resources','generator_vcpus')]:
            d=observation();target=d if section is None else d[section];target.pop(key)
            with self.subTest(missing=key),self.assertRaises(report.EvidenceError):report.summarize(d)
        d=observation();d['measurement']['clock_checked']='yes'
        with self.assertRaises(report.EvidenceError):report.summarize(d)
        d=observation();d['policy']['minimum_expected_alerts']=599
        with self.assertRaisesRegex(report.EvidenceError,'frozen contract'):report.summarize(d)

    def test_review_boundary_and_nearest_rank(self):
        d=observation();r=report.summarize(d)
        self.assertEqual(r['status'],'review_required');self.assertFalse(r['local_run_accepted'])
        self.assertEqual(r['full_run']['alerts']['p99_upper_bound_ns'],500_000_000)
        self.assertEqual(r['full_run']['alerts']['p99_rank'],594)
        self.assertEqual(sum(row['offered_eligible'] for row in d['packet_windows']),108000)
        self.assertEqual(set(r['full_run']['packets_by_phase']),{'sustained','burst_180','burst_480'})
        raw=encoded(d);v=local.reviewed(r,raw,review(d,raw))
        self.assertEqual(v['status'],'reviewed_local_acceptance');self.assertTrue(v['local_run_accepted'])
        self.assertFalse(v['slo_compliance_asserted'])
        for field,value in [('run_id','other'),('observations_sha256','0'*64),('contract_id','local-v2')]:
            bad=review(d,raw);bad[field]=value
            with self.subTest(field=field),self.assertRaises(report.EvidenceError):local.reviewed(r,raw,bad)
        changed=observation();changed['expected_alerts'][0]['durable_ns']=None
        with self.assertRaisesRegex(report.EvidenceError,'summary differs'):
            local.reviewed(r,encoded(changed),review(changed,encoded(changed)))
        for key in local.REVIEW_CHECKS:
            bad=review(d,raw);bad['checks'][key]=False
            with self.subTest(check=key),self.assertRaises(report.EvidenceError):local.reviewed(r,raw,bad)

    def test_missing_and_numeric_failure_survive_invalid_measurement(self):
        for count,kind in [(1,'finite'),(7,'unbounded_missing')]:
            d=observation()
            for a in d['expected_alerts'][:count]:a['durable_ns']=None
            d['integrity']['capture_errors']=1
            r=report.summarize(d)
            self.assertEqual(r['status'],'measurement_failed');self.assertTrue(r['coverage_gaps'])
            self.assertEqual(r['full_run']['alerts']['expected'],600)
            self.assertEqual(r['full_run']['alerts']['missing'],count)
            self.assertEqual(r['full_run']['alerts']['p99_kind'],kind)
            with self.assertRaises(report.EvidenceError):local.reviewed(r,encoded(d),review(d,encoded(d)))
        d=observation()
        for a in d['expected_alerts']:a['durable_ns']+=1
        r=report.summarize(d)
        self.assertEqual(r['status'],'measurement_failed');self.assertEqual(r['full_run']['alerts']['late'],600)
        d['measurement']['clock_uncertainty_ns']=None
        self.assertEqual(report.summarize(d)['status'],'inconclusive')
        # Even without a usable clock bound, a raw latency over 500ms stays failed.
        for a in d['expected_alerts']:a['durable_ns']+=11_000_000
        self.assertEqual(report.summarize(d)['status'],'measurement_failed')

    def test_each_loss_segment_and_exact_one_percent(self):
        for index in (0,3,8):
            for excess,status in ((0,'review_required'),(1,'measurement_failed')):
                d=observation();w=d['packet_windows'][index]
                w['fully_processed']-=w['offered_eligible']//100 + excess if index in (3,8) else 480 + excess
                r=report.summarize(d)
                self.assertEqual(r['status'],status,(index,excess,r['measurement_failures']))
        # An isolated burst failure must not be diluted into the two-burst total.
        d=observation();d['packet_windows'][3]['fully_processed']-=301
        self.assertIn('burst_180: packet loss exceeds 1%',report.summarize(d)['measurement_failures'])

    def test_isolated_burst_latency_and_counter_preservation(self):
        d=observation()
        d['expected_alerts'][180]['durable_ns']+=1
        r=report.summarize(d)
        self.assertEqual(r['full_run']['alerts']['p99_upper_bound_ns'],500_000_000)
        self.assertIn('burst_180: p99 upper bound exceeds 500 ms',r['measurement_failures'])
        d=observation();d['packet_windows'][0]['fully_processed']+=1
        with self.assertRaisesRegex(report.EvidenceError,'exceeds submitted'):report.summarize(d)
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp)/'raw';sources(root)
            first=(root/'offered.jsonl').read_bytes().splitlines(keepends=True)[0]
            with (root/'offered.jsonl').open('ab') as stream:stream.write(first)
            with self.assertRaises(sqlite3.IntegrityError) as caught:adapt(root)
            self.assertIn('UNIQUE constraint',str(caught.exception))
            self.assertFalse((root/'adapter/receipt.json').exists())
            self.assertTrue((root/'adapter/offered.jsonl').read_bytes().endswith(first))

    def test_load_clock_resources_and_incomplete_boundaries(self):
        cases=[('execution_complete',False),('measurement.clock_checked',False),('measurement.clock_uncertainty_ns',None),
               ('measurement.clock_uncertainty_ns',10_000_001),('measurement.live_arrival',False),('measurement.durable_completion',False),
               ('resources.active_rules',99),('resources.sut_vcpus',5),('resources.sut_memory_bytes',3*1024**3+1),
               ('resources.generator_vcpus',3),('resources.generator_memory_bytes',1024**3+1),
               ('integrity.capture_errors',1),('integrity.submission_errors',1),('integrity.raw_evidence_complete',False),
               ('integrity.identity_accounting_complete',False)]
        for path,value in cases:
            d=observation()
            for a in d['expected_alerts']:a['durable_ns']-=100_000_000
            parts=path.split('.');target=d if len(parts)==1 else d[parts[0]];target[parts[-1]]=value
            r=report.summarize(d)
            with self.subTest(path=path):
                self.assertEqual(r['status'],'inconclusive');self.assertFalse(r['local_run_accepted'])
        for count,status in ((5700,'review_required'),(6300,'review_required'),(5699,'inconclusive'),(6301,'inconclusive')):
            d=observation();d['packet_windows'][0].update(offered_eligible=count,submitted_eligible=count,fully_processed=count,offered_bytes=count*512)
            self.assertEqual(report.summarize(d)['status'],status)
        for change in ('short','drain','phase','incomplete','submit','oracle-short','oracle-time'):
            d=observation()
            if change=='short':
                d.update(duration_seconds=540,observed_through_ns=570*NS);d['packet_windows'].pop();d['expected_alerts']=d['expected_alerts'][:540]
            elif change=='drain':d.update(drain_seconds=29,observed_through_ns=629*NS)
            elif change=='phase':d['packet_windows'][3]['phase']='sustained'
            elif change=='incomplete':d['packet_windows'][0]['observation_complete']=False
            elif change=='submit':d['packet_windows'][0]['submitted_eligible']-=1;d['packet_windows'][0]['fully_processed']-=1
            elif change=='oracle-short':d['expected_alerts'].pop()
            else:
                a=d['expected_alerts'][0]
                for key in ('offered_ns','arrival_ns','durable_ns'):a[key]+=NS
            with self.subTest(change=change):self.assertEqual(report.summarize(d)['status'],'inconclusive')

    def test_duplicate_oracle_and_all_anomaly_reasons(self):
        d=observation();d['expected_alerts'].append(copy.deepcopy(d['expected_alerts'][0]))
        with self.assertRaisesRegex(report.EvidenceError,'duplicate expected'):report.summarize(d)
        for reason in ('unexpected','duplicate','after_observation'):
            d=observation();d['anomalies']=[dict(kind='durable',identity='extra',at_ns=631*NS if reason=='after_observation' else NS,reason=reason)]
            r=report.summarize(d)
            self.assertEqual(r['status'],'measurement_failed');self.assertEqual(r['anomalies'],d['anomalies'])


class RetentionTests(unittest.TestCase):
    def test_adapter_retains_missing_extra_duplicate_late_and_failed_submissions(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp)/'raw'
            extra=[dict(kind='durable',event_id='e120',at_ns=631*NS),dict(kind='durable',event_id='extra',at_ns=NS),dict(kind='durable',event_id='e0',at_ns=490_001_000),
                   dict(kind='processed',packet_id='p60',at_ns=631*NS)]
            sources(root,extra=extra)
            offers=[json.loads(line) for line in (root/'offered.jsonl').read_text().splitlines()]
            offers[1]['submitted']=False
            (root/'offered.jsonl').write_bytes(b''.join(json.dumps(row).encode()+b'\n' for row in offers))
            # Remove one expected durable completion. Its oracle remains present.
            rows=[json.loads(line) for line in (root/'events.jsonl').read_text().splitlines()]
            rows=[row for row in rows if row.get('event_id')!='e120' or row['at_ns']>630*NS]
            (root/'events.jsonl').write_bytes(b''.join(json.dumps(row).encode()+b'\n' for row in rows))
            adapt(root)
            doc,raw=report.read_observations(root/'adapter/observations.json')
            self.assertEqual(doc['integrity']['submission_errors'],1)
            self.assertEqual(len(doc['expected_alerts']),10)
            self.assertEqual(sum(w['fully_processed'] for w in doc['packet_windows']),8)
            self.assertEqual({a['reason'] for a in doc['anomalies']},{'unexpected','duplicate','after_observation'})
            self.assertTrue(any(a['identity']=='extra' for a in doc['anomalies']))
            self.assertIsNone(next(a for a in doc['expected_alerts'] if a['event_id']=='e120')['durable_ns'])
            self.assertTrue(any(a['identity']=='e120' and a['reason']=='after_observation' for a in doc['anomalies']))
            self.assertEqual(report.summarize(doc)['status'],'measurement_failed')
            self.assertEqual((root/'events.jsonl').read_bytes(),(root/'adapter/events.jsonl').read_bytes())
            replay=reconstruct.reconstruct(root/'replay',adapter=root/'adapter')
            self.assertEqual(replay['status'],'review_required',replay)
            self.assertTrue(replay['observations_match']);self.assertTrue(replay['observation_bytes_match'])
            self.assertEqual(raw,(root/'replay/rebuilt/observations.json').read_bytes())

    def test_context_contract_binding_and_unknowns(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp);write(root/'observations.json',observation())
            result=retained_context(root,root/'observations.json')
            self.assertEqual(result['status'],'review_required',result)
            self.assertEqual(result['run_identity']['contract_id'],'local-v1')
            context_compare._manifest(result)
            d=json.loads((root/'declaration.json').read_text());d['contract_id']='local-v2'
            write(root/'bad.json',d)
            r=context.retain(root/'bad-context',declaration=root/'bad.json',observations=root/'observations.json',evidence_dir=root/'evidence')
            self.assertEqual(r['status'],'mismatch')
            d=json.loads((root/'declaration.json').read_text());d['hardware']['cpu_model']['value']=None
            write(root/'unknown.json',d)
            r=context.retain(root/'unknown',declaration=root/'unknown.json',observations=root/'observations.json',evidence_dir=root/'evidence')
            self.assertEqual(r['status'],'incomplete')
            # A legacy context cannot bind a local observation even with the same run ID.
            d=json.loads((root/'declaration.json').read_text());d.update(schema_version=1,profile='staging');del d['contract_id'];d['hardware']['deployment_type']['value']='kvm_vm'
            write(root/'legacy.json',d)
            r=context.retain(root/'legacy-context',declaration=root/'legacy.json',observations=root/'observations.json',evidence_dir=root/'evidence')
            self.assertEqual(r['status'],'mismatch')
            # Preserve v1 diagnostic precedence when both kind and profile are bad.
            d.update(artifact_kind='bad',profile='bad')
            with self.assertRaisesRegex(report.EvidenceError,'wrong context kind'):context.validate(d)
            d['artifact_kind']='netsentry_slo_run_context'
            with self.assertRaisesRegex(report.EvidenceError,'invalid profile'):context.validate(d)

    def test_bundle_replay_and_pair_context_contracts(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp);left=root/'left';right=root/'right'
            for folder,run_id,start in ((left,'left-run','2026-10-18T00:00:00Z'),(right,'right-run','2026-10-18T01:00:00Z')):
                r=package(folder,run_id=run_id,start=start)
                self.assertEqual(r['status'],'review_required',r)
                self.assertEqual(r['reported_measurement_status'],'inconclusive')
                retained_context(folder,folder/'adapter/observations.json')
            r=compare.compare(root/'pair',baseline=left/'bundle',candidate=right/'bundle',baseline_context=left/'retained',candidate_context=right/'retained',reconstruct_ledgers=True)
            self.assertEqual(r['status'],'review_required',r)
            self.assertFalse(r['slo_compliance_asserted'])
            self.assertTrue(any(row['field']=='contract_id' and row['equal'] for row in r['condition_comparisons']))
            legacy=root/'legacy';r=package(legacy,profile='staging',run_id='old-run',start='2026-10-18T02:00:00Z')
            self.assertEqual(r['status'],'review_required',r)
            r=compare.compare(root/'cross',baseline=left/'bundle',candidate=legacy/'bundle',reconstruct_ledgers=True)
            self.assertEqual(r['status'],'invalid_evidence',r)
            prod=root/'prod';r=package(prod,profile='prod',run_id='prod-run',start='2026-10-18T03:00:00Z')
            self.assertEqual(r['status'],'review_required',r)
            r=compare.compare(root/'legacy-cross',baseline=legacy/'bundle',candidate=prod/'bundle',reconstruct_ledgers=True)
            self.assertEqual(r['status'],'conditions_differ',r)
            # Mutating only the supplied contract must also fail fresh reconciliation.
            d=json.loads((right/'adapter/manifest.json').read_text());d['contract_id']='local-v2';write(right/'adapter/manifest.json',d)
            r=bundle.reconcile(root/'damaged',adapter=right/'adapter',summary=right/'report.json')
            self.assertEqual(r['status'],'mismatch',r)

    def test_cli_status_review_and_input_preservation(self):
        with tempfile.TemporaryDirectory() as temp:
            root=Path(temp);d=observation();raw=encoded(d);(root/'in.json').write_bytes(raw)
            write(root/'review.json',review(d,raw))
            command=[sys.executable,str(Path(report.__file__)),'--input',str(root/'in.json')]
            result=subprocess.run(command+['--output',str(root/'out.json')],capture_output=True,text=True)
            self.assertEqual(result.returncode,0,result.stderr)
            self.assertEqual(json.loads((root/'out.json').read_text())['status'],'review_required')
            result=subprocess.run(command+['--review',str(root/'review.json'),'--output',str(root/'reviewed.json')],capture_output=True,text=True)
            self.assertEqual(result.returncode,0,result.stderr)
            self.assertEqual(json.loads((root/'reviewed.json').read_text())['status'],'reviewed_local_acceptance')
            # Recompute the reviewed summary from exactly retained bytes.
            retained=bundle._Bundle(root,64*1024**2);(root/'adapter').mkdir();(root/'adapter/observations.json').write_bytes(raw)
            retained.inventory['adapter/observations.json']=dict(file='adapter/observations.json',bytes=len(raw),sha256=hashlib.sha256(raw).hexdigest(),complete=True)
            bundle._summary(retained,json.loads((root/'reviewed.json').read_text()),d)
            for mode,code in (('fail',1),('invalid',3),('bad',2)):
                changed=observation()
                if mode=='fail':changed['expected_alerts'][0]['durable_ns']=None
                if mode=='invalid':changed['measurement']['clock_uncertainty_ns']=None
                if mode=='bad':changed['contract_id']='local-v2'
                write(root/'in.json',changed)
                result=subprocess.run(command+['--output',str(root/(mode+'.json'))],capture_output=True,text=True)
                self.assertEqual(result.returncode,code,result.stderr)
            # Existing output is never replaced, and stale review cannot accept new input.
            before=(root/'out.json').read_bytes()
            result=subprocess.run(command+['--review',str(root/'review.json'),'--output',str(root/'out.json')],capture_output=True,text=True)
            self.assertEqual(result.returncode,2);self.assertEqual(before,(root/'out.json').read_bytes())
            (root/'in.json').write_bytes(raw)
            with mock.patch.object(Path,'home',return_value=root),contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(report.main(['--input',str(root/'in.json')]),0)
            self.assertTrue((root/'benchmarks/local-v1/synthetic-run/report.json').exists())
            self.assertEqual((root/'in.json').read_bytes(),raw)


if __name__ == '__main__':
    unittest.main()
