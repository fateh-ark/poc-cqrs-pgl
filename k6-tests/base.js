import http from 'k6/http';
import { Faker } from 'k6/x/faker';
import { check, group, sleep } from 'k6';

let faker = new Faker(11)

// --- AUTH CONFIG ---
const KEYCLOAK_HOST = __ENV.KEYCLOAK_HOST || 'http://108.142.59.225:8080/';
const CLIENT_ID = __ENV.CLIENT_ID || 'pcg-client';
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'UG0J9b2KpNGtf0i3qNTeeSc36hwqZhKz';
const USERNAME = __ENV.KC_USER || 'writeradmin';
const PASSWORD = __ENV.KC_PASS || '12345';
const CLIENT_SCOPE = 'openid profile email';
// --- END AUTH CONFIG ---

const BASE_URL = __ENV.BASE_URL || 'http://108.142.26.234:4180/';

export function authenticate() {
  const url = `${KEYCLOAK_HOST}/realms/pcg/protocol/openid-connect/token`;
  const payload = {
    grant_type: 'password',
    client_id: CLIENT_ID,
    username: USERNAME,
    password: PASSWORD,
    client_secret: CLIENT_SECRET,
    scope: CLIENT_SCOPE,
  };
  const params = {
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
  };
  const res = http.post(url, Object.entries(payload).map(([k,v])=>`${k}=${encodeURIComponent(v)}`).join('&'), params);
  check(res, { 'token status is 200': (r) => r.status === 200 });
  const json = res.json();
  return json['access_token'];
}

export function postEndpoint(bearerToken) {
  group('Post endpoint', function () {
    if (!bearerToken) bearerToken = authenticate();
    const bookData = {
        title:  faker.book.bookTitle(),
        author: faker.book.bookAuthor()
    }
    const res = http.post(
        BASE_URL + 'books',
        JSON.stringify(bookData),
        {headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${bearerToken}` },}
    );
    const cek = check(res, { 'status is 201': (r) => r.status === 201 });
    sleep(1);
    return cek
  });
}

export function getByIdEndpoint(bearerToken) {
  group('Get by ID endpoint', function () {
    if (!bearerToken) bearerToken = authenticate();
    const randomId = Math.floor(Math.random() * 20) + 1;
    const url = BASE_URL + `books/${randomId}`;
    const res = http.get(url, {headers: { 'Authorization': `Bearer ${bearerToken}` }});
    const cek = check(res, { 'status is 200 or 404': (r) => r.status === 200 || r.status === 404 });
    sleep(1);
    return cek
  });
}