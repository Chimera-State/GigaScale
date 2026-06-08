import http from 'k6/http';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const status200 = new Counter('Status_200_Success');
const status409 = new Counter('Status_409_Conflict');
const status429 = new Counter('Status_429_RateLimit');
const status500 = new Counter('Status_500_ServerError');

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
    amount: 100.50,
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

  responses.forEach((res) => {
    if (res.status === 200) status200.add(1);
    else if (res.status === 409) status409.add(1);
    else if (res.status === 429) status429.add(1);
    else if (res.status >= 500) status500.add(1);
  });
}
