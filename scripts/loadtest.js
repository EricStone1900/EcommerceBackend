import http from 'k6/http';
import { check, sleep, group } from 'k6';

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test user credentials
const TEST_USER = {
  email: 'loadtest@example.com',
  password: 'LoadTest123!',
};

// A fixed product ID for detail endpoint testing (seeded in setup)
let PRODUCT_ID = 1;

export const options = {
  vus: 50,
  duration: '1m',
  thresholds: {
    http_req_duration: ['p(95)<2000', 'p(99)<5000'],
    http_req_failed: ['rate<0.05'],
  },
};

function getToken() {
  const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  if (loginRes.status !== 200) {
    // Try to register first
    http.post(`${BASE_URL}/api/v1/auth/register`,
      JSON.stringify({
        email: TEST_USER.email,
        password: TEST_USER.password,
        name: 'Load Test User',
      }),
      { headers: { 'Content-Type': 'application/json' } }
    );

    const retryRes = http.post(`${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
      { headers: { 'Content-Type': 'application/json' } }
    );

    if (retryRes.status !== 200) return '';

    try {
      return JSON.parse(retryRes.body).data.access_token;
    } catch (e) {
      return '';
    }
  }

  try {
    return JSON.parse(loginRes.body).data.access_token;
  } catch (e) {
    return '';
  }
}

export function setup() {
  const token = getToken();

  // If we got a token, try to find an existing product via the list endpoint
  let productID = 0;

  if (token) {
    const listRes = http.get(`${BASE_URL}/api/v1/products?page=1&page_size=5`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });

    if (listRes.status === 200) {
      try {
        const body = JSON.parse(listRes.body);
        if (body.data && body.data.list && body.data.list.length > 0) {
          productID = body.data.list[0].id;
        }
      } catch (e) {}
    }
  }

  console.log(`setup complete: token=${token ? 'yes' : 'no'}, productID=${productID}`);

  return {
    token: token,
    productID: productID,
  };
}

export default function (data) {
  const token = data.token;
  const prodID = data.productID || 1;

  // Group 1: Login
  group('Authentication', function () {
    const loginRes = http.post(`${BASE_URL}/api/v1/auth/login`,
      JSON.stringify({ email: TEST_USER.email, password: TEST_USER.password }),
      { headers: { 'Content-Type': 'application/json' } }
    );

    check(loginRes, {
      'login status is 200': (r) => r.status === 200,
      'login has access_token': (r) => {
        try { return JSON.parse(r.body).data.access_token !== ''; } catch (e) { return false; }
      },
    });
  });

  sleep(1);

  // Group 2: Product List
  if (token) {
    group('Product List', function () {
      const listRes = http.get(`${BASE_URL}/api/v1/products?page=1&page_size=20`, {
        headers: { 'Authorization': `Bearer ${token}` },
      });

      check(listRes, {
        'product list status is 200': (r) => r.status === 200,
        'product list has data': (r) => {
          try {
            const body = JSON.parse(r.body);
            return body.data && body.data.list !== undefined;
          } catch (e) { return false; }
        },
      });
    });
  }

  sleep(1);

  // Group 3: Product Detail (public endpoint, no auth needed)
  group('Product Detail', function () {
    const detailRes = http.get(`${BASE_URL}/api/v1/products/${prodID}`);

    check(detailRes, {
      'product detail status is 200_or_404': (r) => r.status === 200 || r.status === 404,
    });

    if (detailRes.status === 200) {
      check(detailRes, {
        'product detail returns data': (r) => {
          try { return JSON.parse(r.body).data !== null; } catch (e) { return false; }
        },
      });
    }
  });

  sleep(1);
}
