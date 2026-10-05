import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';
import {
  TokenSession,
  observe,
  requestOutage,
  sourceStatus,
} from './auth.js';

const outageSeconds = Number(__ENV.OUTAGE_SECONDS || 20);
const recoverySeconds = Number(__ENV.RECOVERY_SECONDS || 30);
const outageVUs = Number(__ENV.OUTAGE_VUS || 5);
const recoveryVUs = Number(__ENV.RECOVERY_VUS || 5);
const pause = Number(__ENV.SLEEP || 0.2);
const outageStatusObserved = new Rate('outage_status_observed');
const recoveryAvailable = new Rate('recovery_available');
const outageSession = new TokenSession();
const recoverySession = new TokenSession();

export const options = {
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(90)', 'p(95)', 'p(99)'],
  scenarios: {
    outage: {
      executor: 'constant-vus',
      vus: outageVUs,
      duration: `${outageSeconds}s`,
      exec: 'duringOutage',
      tags: { scenario: 'outage', phase: 'outage' },
    },
    restore: {
      executor: 'shared-iterations',
      vus: 1,
      iterations: 1,
      startTime: `${outageSeconds + 1}s`,
      exec: 'restorePVMBG',
      tags: { scenario: 'outage', phase: 'restore' },
    },
    recovery: {
      executor: 'constant-vus',
      vus: recoveryVUs,
      duration: `${recoverySeconds}s`,
      startTime: `${outageSeconds + 2}s`,
      exec: 'duringRecovery',
      tags: { scenario: 'outage', phase: 'recovery' },
    },
  },
  thresholds: {
    outage_status_observed: ['rate>0'],
    recovery_available: ['rate>0.5'],
    system_error_rate: ['rate<0.01'],
  },
};

export function setup() {
  requestOutage(true, 'error');
}

export function duringOutage() {
  const response = outageSession.request('GET', '/v1/hazards?limit=20', {
    tags: { endpoint: 'hazards' },
  });
  observe(response, { endpoint: 'hazards', phase: 'outage' }, [200, 503]);
  let body = null;
  try {
    body = response.json();
  } catch (_) {
    body = null;
  }
  const pvmbg = sourceStatus(body, 'PVMBG');
  const visible = response.status === 503 || (pvmbg && (pvmbg.status !== 'HEALTHY' || pvmbg.stale_since));
  outageStatusObserved.add(visible ? 1 : 0);
  check(response, {
    'outage request returns an expected status': (result) => result.status === 200 || result.status === 429 || result.status === 503,
  });
  sleep(pause);
}

export function restorePVMBG() {
  requestOutage(false);
}

export function duringRecovery() {
  const response = recoverySession.request('GET', '/v1/hazards?limit=20', {
    tags: { endpoint: 'hazards' },
  });
  observe(response, { endpoint: 'hazards', phase: 'recovery' }, [200, 503]);
  let body = null;
  try {
    body = response.json();
  } catch (_) {
    body = null;
  }
  const pvmbg = sourceStatus(body, 'PVMBG');
  const available = response.status === 200 && pvmbg && pvmbg.status === 'HEALTHY';
  recoveryAvailable.add(available ? 1 : 0);
  check(response, {
    'recovery request returns an expected status': (result) => result.status === 200 || result.status === 429 || result.status === 503,
  });
  sleep(pause);
}

export function teardown() {
  requestOutage(false);
}
