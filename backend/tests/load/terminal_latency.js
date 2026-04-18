import ws from 'k6/ws';
import { check } from 'k6';

export let options = {
  vus: 1,
  duration: '10s',
};

export default function () {
  const url = 'ws://localhost:8080/api/vms/test-id/terminal?token=TEST_TOKEN';

  const res = ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      // SC-003: Terminal keystroke latency check
      const start = Date.now();
      socket.send(JSON.stringify({ type: 'input', data: 'ls\n' }));
      
      socket.on('message', (data) => {
        const end = Date.now();
        console.log(`Latency: ${end - start}ms`);
      });
    });

    socket.setTimeout(() => {
      socket.close();
    }, 1000);
  });

  check(res, { 'Connected successfully': (r) => r && r.status === 101 });
}
