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
    volcanic: {
      executor: 'constant-vus', vus, duration,
      exec: 'volcanic', tags: { scenario: 'volcanic-concurrent' },
    },
  },
  thresholds: {
    'business_latency{endpoint:seismic}': ['p(95)<300'],
    system_error_rate: ['rate<0.01'],
    auth_error_rate: ['rate<0.01'],
    'successful_requests{endpoint:seismic}': ['count>0'],
    'successful_requests{endpoint:volcanic}': ['count>0'],
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

export function volcanic() {
  const response = session.request('GET', '/v1/hazards/volcanic?limit=20', {tags:{endpoint:'volcanic'}});
  observe(response,{endpoint:'volcanic'},[200]);
  sleep(pause);
}
