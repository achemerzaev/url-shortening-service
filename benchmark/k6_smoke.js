import http, { head } from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  duration: '5s',
};

export default function () {
  const name = `user__${__VU}_${__ITER}`;
  const email = `user__${__VU}_${__ITER}$@test.com`;
  const password = '12345678';

  // 1. register
  http.post('http://localhost:8080/register', JSON.stringify({
    name,
    email,
    password,
  }), { headers: { 'Content-Type': 'application/json' } });


  // 2. login
  const loginRes = http.post('http://localhost:8080/login', JSON.stringify({
    email,
    password,
  }), { headers: { 'Content-Type': 'application/json' } });

  const token = loginRes.json('access_token');

  const headers = {
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  // 3. create short url

  const createRes = http.post(
    'http://localhost:8080/shorten',
    JSON.stringify({ url: 'https://google.com' }),
    headers
  );

  const code = createRes.json('shortcode')

  // 4. severel "get"s
  for (let i = 0; i < 3; i++) {
    http.get(`http://localhost:8080/shorten/${code}`, {
      ...headers,
      redirects: 0,
    } );
  }

  // 5. stats
  http.get(`http://localhost:8080/shorten/${code}/stats`, headers);

  // 6. delete
  http.del(`http://localhost:8080/shorten/${code}`, null, headers);

  sleep(1);

}


// curl -X POST http://localhost:8080/shorten \
//      -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzc1Mjk3MjYsImp0aSI6ImUyZTM4ZDc4YzkzNzkwZDA4NDc2MTlmZWIxNGE3OThhIiwidXNlcl9pZCI6MX0.qW3xEOWoCQUdimx0b4lvrpN40n8NGhu1_K7dB-sYslI" \
//      -H "Content-Type: application/json" \
//      -d '{"url": "https://google.com"}'
