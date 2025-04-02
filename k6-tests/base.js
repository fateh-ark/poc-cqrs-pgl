import http from 'k6/http';
import { Faker } from 'k6/x/faker';
import { check, group, sleep } from 'k6';

let faker = new Faker(11)

export function postEndpoint() {
  group('Post endpoint', function () {
    const bookData = {
        title:  faker.book.bookTitle(),
        author: faker.book.bookAuthor()
    }
    const res = http.post(
        'http://localhost:80/books',
        JSON.stringify(bookData),
        {headers: { 'Content-Type': 'application/json' },}
    );
    const cek = check(res, { 'status is 201': (r) => r.status === 201 });
    sleep(1);
    return cek
  });
}

export function getByIdEndpoint() {
  group('Get by ID endpoint', function () {
    const randomId = Math.floor(Math.random() * 20) + 1;
    const url = `http://localhost:80/books/${randomId}`;
    const res = http.get(url);
    const cek = check(res, { 'status is 200 or 404': (r) => r.status === 200 || r.status === 404 });
    sleep(1);
    return cek
  });
}