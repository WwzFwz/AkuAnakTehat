import { check, sleep } from 'k6';
import { TokenSession, observe } from './auth.js';

const vus = Number(__ENV.VUS || 50);
const duration = __ENV.DURATION || '90s';
const pause = Number(__ENV.SLEEP || 0.1);
const session = new TokenSession();

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
  scenarios: {
    sustained: {
      executor: 'constant-vus',
      vus,
      duration,
      tags: { scenario: 'sustained' },
    },
  },
  thresholds: {
    system_error_rate: ['rate<0.01'],
    auth_error_rate: ['rate<0.01'],
    successful_requests: ['count>0'],
    'controlled_429': ['count>=0'],
  },
};

export default function () {
  const type = __VU % 2 === 0 ? 'VOLCANIC' : 'SEISMIC';
  const endpoint = type === 'VOLCANIC' ? 'volcanic' : 'seismic';
  const response = session.request('GET', `/v1/hazards/${endpoint}?limit=20`, {
    tags: { endpoint },
  });
  observe(response, { endpoint }, [200]);
  check(response, {
    'sustained request returns data or controlled rejection': (result) => result.status === 200 || result.status === 429,
  });
  sleep(pause);
}
