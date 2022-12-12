import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: 2,
  duration: '10s',
};

export default function () {
  // http.post('http://localhost:5555/message/v1/publish', JSON.stringify({
  http.post('https://xmtp.snormore.dev:5555/message/v1/publish', JSON.stringify({
    envelopes: [{
      contentTopic: 'topic',
      timestampNs: 1
    }]
  }));
  sleep(1);
}
