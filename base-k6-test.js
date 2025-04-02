import http from 'k6/http';
import { check, group, sleep } from 'k6';

// Endpoint 1: /users
export function getAllEndpoint() {
  group('Get All Books endpoint', function () {
    const res = http.get('http://localhost:80/books');
    check(res, { 'status is 200': (r) => r.status === 200 });
    sleep(1);
  });
}

// Endpoint 2: /products
export function getByIdEndpoint() {
  group('Read random book endpoint', function () {
    const randomId = Math.floor(Math.random() * 5) + 1;
    const url = `http://localhost:80/books/${randomId}`;
    const res = http.get(url);
    check(res, { 'status is 200': (r) => r.status === 200 });
    sleep(1);
  });
}

export const options = {
  stages: [
    { duration: '10s', target: 2 }, // Ramp up to 5 VUs over 10 seconds
    { duration: '20s', target: 5 }, // Ramp up to 5 VUs over 10 seconds
    { duration: '10s', target: 10 }, // Maintain 10 VUs for 20 seconds
    { duration: '5s', target: 50 }, // Maintain 10 VUs for 20 seconds
    // { duration: '10s', target: 0 }, // Ramp down to 0 VUs over 10 seconds
    // { duration: '20s', target: 100 }, // Maintain 10 VUs for 20 seconds
    // { duration: '5s', target: 200 }, // Maintain 10 VUs for 20 seconds
    { duration: '5s', target: 50 }, // Maintain 10 VUs for 20 seconds
    { duration: '20s', target: 5 }, // Ramp up to 5 VUs over 10 seconds
    { duration: '10s', target: 2 }, // Ramp up to 5 VUs over 10 seconds

  ],
};

export default function () {
  getAllEndpoint();
  getByIdEndpoint();
}