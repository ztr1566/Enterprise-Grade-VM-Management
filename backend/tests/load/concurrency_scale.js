import http from 'k6/http';
import ws from 'k6/ws';
import { check, sleep } from 'k6';

export let options = {
  vus: 50, // SC-005: 50 concurrent target VM monitoring sessions
  duration: '30s',
};

export default function () {
  // 1. Test REST API Scaling (Polling Stats)
  const statsRes = http.get('http://localhost:8080/api/vms/test-id/stats', {
    headers: { 'Authorization': 'Bearer TEST_TOKEN' }
  });
  check(statsRes, { 'Stats status is 200': (r) => r.status === 200 });

  // 2. Test WebSocket Scaling (Terminal)
  const url = 'ws://localhost:8080/api/vms/test-id/terminal?token=TEST_TOKEN';
  ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      socket.send(JSON.stringify({ type: 'input', data: 'uptime\n' }));
      socket.close();
    });
  });

  sleep(1);
}
