import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  vus: 1,
  iterations: 1,
};

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

export default function () {
  const url = 'http://gateway:8080/api/v1/reserve';

  const sharedUserId = generateAlphanumID();
  const sharedTripId = generateAlphanumID();
  const sharedSeatId = generateAlphanumID();
  const sharedIdempotencyKey = uuidv4();

  const payload = JSON.stringify({
    user_id: sharedUserId,
    trip_id: sharedTripId,
    seat_id: sharedSeatId,
    idempotency_key: sharedIdempotencyKey,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
  };

  const requests = [];
  for (let i = 0; i < 10; i++) {
    requests.push({
      method: 'POST',
      url: url,
      body: payload,
      params: params,
    });
  }

  const responses = http.batch(requests);

  let successCount = 0;
  let conflictCount = 0;

  responses.forEach((res) => {
    if (res.status === 200 || res.status === 201) {
      successCount++;
    } else if (res.status === 409) {
      conflictCount++;
    }
  });

  check(responses, {
    'Sadece 1 istek basarili oldu (200/201)': () => successCount === 1,
    'Tam 9 istek zaten islendi (409) dondu': () => conflictCount === 9,
  });
}