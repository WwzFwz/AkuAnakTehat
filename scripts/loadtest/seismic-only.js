import { check, sleep } from 'k6';
import { TokenSession, observe } from './auth.js';

const vus = Number(__ENV.VUS || 10);
const duration = __ENV.DURATION || '60s';
const pause = Number(__ENV.SLEEP || 0.1);
const session = new TokenSession();

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
  scenarios: {
    seismic: {
      executor: 'constant-vus',
      vus,
      duration,
      tags: { scenario: 'seismic-only' },
    },
  },
  thresholds: {
    'business_latency{endpoint:seismic}': ['p(95)<300'],
    system_error_rate: ['rate<0.01'],
    'controlled_429': ['count>=0'],
  },
};

export default function () {
  const response = session.request('GET', '/v1/hazards/seismic?limit=20', {
    tags: { endpoint: 'seismic' },
  });
  observe(response, { endpoint: 'seismic' }, [200]);
  check(response, {
    'seismic returns data or controlled rejection': (result) => result.status === 200 || result.status === 429,
  });
  sleep(pause);
}
