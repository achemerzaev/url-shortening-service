import http, { head } from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    constant_load: {
      executor: 'constant-arrival-rate',
      rate: 1000,
      timeUnit: '1s',
      duration: '60s',
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
  },
};

const BASE = 'http://localhost:8080'

export function setup() {
  const name = 'John';
  const email = `bench1@test.com`;
  const password = '12345678';

  const resp = http.post( `${BASE}/register`, JSON.stringify({
    name,
    email,
    password,
  }), { 
    headers: { 'Content-Type': 'application/json' } ,
    tags: { name: 'POST /register' },
  });

  const token = resp.json('access_token');

  return { token };
}

export default function (data) {
  const params = {
    headers: {
      Authorization: `Bearer ${data.token}`,
      'Content-Type': 'application/json',
    },
  };

  // 1. create short url

  const createRes = http.post(
    `${BASE}/shorten`,
    JSON.stringify({ url: 'https://google.com' }),
    {
      ...params,
      tags: { name: 'POST /shorten' },
    }
  );

  const code = createRes.json('shortcode')

  // 2. severel "get"s
  for (let i = 0; i < 3; i++) {
    http.get(`${BASE}/shorten/${code}`, {
      ...params,
      redirects: 0,
      tags: { name: 'GET /shorten/:code' },
    } );
  }


  // 3. stats
  http.get(`${BASE}/shorten/${code}/stats`,
    {
      ...params,
      tags: { name: 'GET /shorten/:code/stats' },
    }
  );

  // 4. delete
  http.del(`${BASE}/shorten/${code}`, 
    null, 
    {
      ...params,
      tags: { name: 'DELETE /shorten/:code' },
    }
  );


  sleep(1);

}