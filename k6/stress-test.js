import http from 'k6/http';
import { sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const status200 = new Counter('Status_200_Success');
const status409 = new Counter('Status_409_Conflict');
const status429 = new Counter('Status_429_RateLimit');
const status500 = new Counter('Status_500_ServerError');

export const options = {
    stages: [
        { duration: '2m', target: 10 },
        { duration: '2m', target: 50 },
        { duration: '2m', target: 100 },
        { duration: '2m', target: 200 },
    ],
};

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

export default function () {
    const url = 'http://gateway:8080/api/v1/reserve';

    const payload = JSON.stringify({
        user_id: generateAlphanumID(),
        trip_id: generateAlphanumID(),
        seat_id: generateAlphanumID(),
        idempotency_key: uuidv4(),
        amount: 100.50,
    });
    const params = {
        headers: { 'Content-Type': 'application/json' },
    };
    
    const res = http.post(url, payload, params);

    if (res.status === 200) status200.add(1);
    else if (res.status === 409) status409.add(1);
    else if (res.status === 429) status429.add(1);
    else if (res.status >= 500) status500.add(1);

    sleep(1);
}