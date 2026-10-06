import http from 'k6/http';
import { sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

export const API_URL = (__ENV.API_URL || 'http://127.0.0.1:8080').replace(/\/+$/, '');
export const AUTH_URL = (__ENV.AUTH_URL || 'http://127.0.0.1:8090').replace(/\/+$/, '');
export const PVMBG_URL = (__ENV.PVMBG_URL || 'http://127.0.0.1:8082').replace(/\/+$/, '');
export const HTTP_TIMEOUT = __ENV.K6_HTTP_TIMEOUT || '5s';

export const controlled429 = new Counter('controlled_429');
export const auth429 = new Counter('auth_429');
export const systemErrorRate = new Rate('system_error_rate');
export const businessLatency = new Trend('business_latency', true);
export const authErrorRate = new Rate('auth_error_rate');
export const businessRequests = new Counter('business_requests');
export const successfulRequests = new Counter('successful_requests');
export const expiredAccessRetries = new Counter('expired_access_retries');
const credentials = __ENV.CLIENTS_FILE ? JSON.parse(open(__ENV.CLIENTS_FILE)) : [];

function encodeForm(values) {
  return Object.keys(values)
    .map((key) => `${encodeURIComponent(key)}=${encodeURIComponent(values[key])}`)
    .join('&');
}

function tokenParams() {
  return {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    timeout: HTTP_TIMEOUT,
    tags: { name: 'oauth_token', kind: 'auth' },
  };
}

function updateToken(session, response) {
  if (response.status !== 200) {
    throw new Error(`token request failed with status ${response.status}`);
  }
  const body = response.json();
  if (!body.access_token || !body.refresh_token || !body.expires_in) {
    throw new Error('token response is incomplete');
  }
  session.accessToken = body.access_token;
  session.refreshToken = body.refresh_token;
  session.expiresAt = Date.now() + Number(body.expires_in) * 1000;
  const refreshWindow = Math.min(5000, Math.max(0, session.expiresAt - Date.now() - 5000));
  session.refreshAt = session.expiresAt - 5000 - Math.random() * refreshWindow;
  return session.accessToken;
}

export class TokenSession {
  constructor() {
    this.clientID = __ENV.FIELD_CLI_CLIENT_ID || __ENV.CLIENT_ID || 'field-team';
    this.clientSecret = __ENV.FIELD_CLI_CLIENT_SECRET || __ENV.CLIENT_SECRET || '';
    if (!this.clientSecret) {
      const credential = credentials.find((entry) => entry.client_id === this.clientID);
      this.clientSecret = credential ? credential.client_secret : '';
    }
    this.accessToken = '';
    this.refreshToken = '';
    this.expiresAt = 0;
    this.refreshAt = 0;
  }

  token() {
    if (!this.clientSecret) {
      throw new Error('FIELD_CLI_CLIENT_SECRET is required');
    }
    if (this.accessToken && Date.now() < this.refreshAt) {
      return this.accessToken;
    }

    if (this.refreshToken) {
      const response = this.exchange({ grant_type: 'refresh_token', refresh_token: this.refreshToken });
      if (response.status === 200) {
        return updateToken(this, response);
      }
      if (response.status !== 400) {
        throw new Error(`token refresh failed with status ${response.status}`);
      }
      this.accessToken = '';
      this.refreshToken = '';
    }

    const response = this.exchange({
      grant_type: 'client_credentials',
      client_id: this.clientID,
      client_secret: this.clientSecret,
    });
    return updateToken(this, response);
  }

  exchange(values) {
    let response;
    for (let attempt = 0; attempt < 4; attempt += 1) {
      response = http.post(`${AUTH_URL}/oauth/token`, encodeForm(values), tokenParams());
      if (response.status !== 429) {
        authErrorRate.add(response.status !== 200, { kind: 'auth' });
        return response;
      }
      auth429.add(1);
      if (attempt < 3) {
        sleep(1);
      }
    }
    authErrorRate.add(1, { kind: 'auth' });
    return response;
  }

  invalidate() {
    this.accessToken = '';
    this.expiresAt = 0;
    this.refreshAt = 0;
  }

  request(method, path, params = {}) {
    let token = this.token();
    const correlationID = `field-k6-${__VU}-${__ITER}-${Date.now()}`;
    let response;
    for (let attempt = 0; attempt < 2; attempt += 1) {
      const headers = Object.assign({}, params.headers || {}, {
        Accept: 'application/json',
        Authorization: `Bearer ${token}`,
        'X-Correlation-ID': correlationID,
      });
      const requestParams = Object.assign({}, params, {
        headers,
        timeout: params.timeout || HTTP_TIMEOUT,
        tags: Object.assign({}, params.tags || {}, { kind: 'business' }),
      });
      response = http.request(method, `${API_URL}${path}`, null, requestParams);
      if (response.status !== 401 || attempt === 1) {
        return response;
      }
      expiredAccessRetries.add(1);
      this.invalidate();
      token = this.token();
    }
    return response;
  }
}

export function observe(response, tags = {}, expectedStatuses = [200]) {
  const status = response.status;
  const labels = Object.assign({}, tags);
  controlled429.add(status === 429 ? 1 : 0, labels);
  businessRequests.add(1, labels);
  // Exclude controlled rejection from BOTH numerator and denominator.
  if (status !== 429) systemErrorRate.add(expectedStatuses.indexOf(status) === -1, labels);
  successfulRequests.add(status === 200 ? 1 : 0, labels);
  // Fast rejection must not make the successful-request p95 look better.
  if (status === 200) businessLatency.add(response.timings.duration, labels);
}

export function sourceStatus(body, source) {
  if (!body || !Array.isArray(body.sources)) {
    return null;
  }
  for (const item of body.sources) {
    if (item.source === source) {
      return item;
    }
  }
  return null;
}

export function requestOutage(enabled, mode = 'error') {
  const adminKey = __ENV.PVMBG_ADMIN_KEY || '';
  if (!adminKey) {
    throw new Error('PVMBG_ADMIN_KEY is required for the outage scenario');
  }
  const response = http.post(
    `${PVMBG_URL}/admin/outage`,
    JSON.stringify({ enabled, mode }),
    {
      headers: {
        'Content-Type': 'application/json',
        'X-Admin-Key': adminKey,
      },
      timeout: HTTP_TIMEOUT,
      tags: { name: 'pvmbg_admin_outage', kind: 'control' },
    },
  );
  if (response.status !== 200) {
    throw new Error(`PVMBG outage control failed with status ${response.status}`);
  }
  return response;
}
