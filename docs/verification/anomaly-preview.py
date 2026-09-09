"""Loopback-only UI fixture; no production credentials or model requests.
Run from the source root after building web/dist, then open localhost:3340.
"""
import json
import time
import uuid
from http.server import ThreadingHTTPServer, SimpleHTTPRequestHandler
from pathlib import Path
from urllib.parse import urlparse, parse_qs

ROOT = Path(__file__).resolve().parents[2] / 'web' / 'dist'
policy = dict(enabled=True, window_minutes=5, schema_threshold=10, rate_threshold=100, upstream_threshold=5, version=0)
event = dict(id='a'*32, user_id=305, channel_id=1, model='gpt-6-astra', kind='invalid_schema', count=12, threshold=10, first_seen=int(time.time())-220, last_seen=int(time.time())-5, alert=True, acknowledged_by=0, acknowledged_at=0, acknowledged_count=0)
audit = []
class Handler(SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=str(ROOT), **kwargs)
    def log_message(self, *args):
        pass
    def respond(self, data, status=200):
        raw=json.dumps(dict(success=status<400, data=data)).encode()
        self.send_response(status); self.send_header('Content-Type','application/json'); self.send_header('Content-Length',str(len(raw))); self.end_headers();self.wfile.write(raw)
    def do_GET(self):
        p=urlparse(self.path); q=parse_qs(p.query)
        if p.path=='/api/setup': return self.respond(dict(status=True))
        if p.path=='/api/status': return self.respond(dict(system_name='冰山公益站 · 本地预览',version='anomaly-preview',setup=True))
        if p.path=='/api/user/self/abuse': return self.respond(dict(suspended=False,blocked_until=0))
        if p.path=='/api/anomalies/settings': return self.respond(policy)
        if p.path=='/api/anomalies/audit': return self.respond(audit)
        if p.path=='/api/anomalies/events':
            matched=not q.get('user_id') or q['user_id'][0]=='305'
            matched=matched and (not q.get('kind') or q['kind'][0]==event['kind'])
            if q.get('state')==['pending'] and event['acknowledged_count']>=event['count']: matched=False
            return self.respond(dict(items=[event] if matched else [],total=int(matched),pending=int(event['acknowledged_count']<event['count']),retained=1))
        if p.path=='/api/abuse/settings': return self.respond(dict(mode='observe',limit_10m=3,limit_24h=8,freeze_minutes=60,enabled_rules=[],rules=[],evidence_capture_ready=False))
        if p.path.startswith('/api/abuse/'): return self.respond(dict(items=[],total=0))
        if p.path.startswith('/api/'): return self.respond('')
        if not (ROOT / p.path.lstrip('/')).is_file(): self.path='/index.html'
        return super().do_GET()
    def do_POST(self):
        global event
        if self.path=='/api/user/auth/refresh':
            return self.respond(dict(access_token='local-preview-only',token_type='Bearer',access_expires_at=int(time.time())+3600,user=dict(id=1,username='preview-root',display_name='本地演示',role=100,status=1,group='default',quota=0,permissions=dict(sidebar_settings=False)),session=dict(sid='preview',current=True,login_method='fixture',ip='127.0.0.1',user_agent='local-preview',created_at=int(time.time()),last_active_at=int(time.time()),expires_at=int(time.time())+3600)))
        if self.path.endswith('/acknowledge'):
            self.rfile.read(int(self.headers.get('Content-Length',0)))
            event.update(acknowledged_count=event['count'],acknowledged_by=1,acknowledged_at=int(time.time()))
            audit.insert(0,dict(id=uuid.uuid4().hex,operator_id=1,action='acknowledged',event_id=event['id'],created_at=int(time.time())))
            return self.respond({})
        return self.respond({})
    def do_PUT(self):
        global policy
        if self.path=='/api/anomalies/settings':
            policy=json.loads(self.rfile.read(int(self.headers['Content-Length'])))
            policy['version']+=1
            audit.insert(0,dict(id=uuid.uuid4().hex,operator_id=1,action='settings_updated',created_at=int(time.time()),policy=policy.copy()))
            return self.respond(policy)
        self.respond({},404)
if __name__=='__main__':
    print('Local synthetic preview: http://127.0.0.1:3340/anomalies',flush=True)
    ThreadingHTTPServer(('127.0.0.1',3340),Handler).serve_forever()
