import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    stages: [
      { duration: '10s', target: 100 },
    ],
  };

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

export default function () {
    const url = 'http://gateway:8080/api/v1/reserve';

    const payload = JSON.stringify({
        user_id:   generateAlphanumID(),
        trip_id:   generateAlphanumID(),
        seat_id:   generateAlphanumID(),
        idempotency_key:  uuidv4(),
    });
    const params = {
        headers: { 'Content-Type': 'application/json' },
    };
    const res = http.post(url, payload, params);

    check(res, { 
        'status 429 veya 200': (r) => r.status === 429 || r.status === 200,  });



}