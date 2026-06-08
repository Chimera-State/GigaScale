import http, { expectedStatuses, setResponseCallback } from 'k6/http';
import { sleep } from 'k6';
import { Counter } from 'k6/metrics';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const status200 = new Counter('Status_200_Success');
const status409 = new Counter('Status_409_Conflict');
const status429 = new Counter('Status_429_RateLimit');
const status500 = new Counter('Status_500_ServerError');

setResponseCallback(expectedStatuses(200, 409, 429));

export const options = {
    insecureSkipTLSVerify: true,
    stages: [
        { duration: '5s', target: 2000 },
        { duration: '10s', target: 10000 },
        { duration: '5s', target: 0 },
    ],
};

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

function generateRandomIP(vuId) {
    const ipPoolIndex = vuId % 200;
    return `192.168.100.${ipPoolIndex}`;
}

const vusIPs = {};

export default function () {
    if (!vusIPs[__VU]) {
        vusIPs[__VU] = generateRandomIP(__VU);
    }

    const myIp = vusIPs[__VU];
    const url = 'http://gateway:8080/api/v1/reserve';

    const hotSeats = ["12A", "12B", "12C"];
    const selectedSeat = hotSeats[Math.floor(Math.random() * hotSeats.length)];

    const payload = JSON.stringify({
        user_id: generateAlphanumID(),
        trip_id: "550e8400-e29b-41d4-a716-446655440000",
        seat_id: selectedSeat,
        idempotency_key: uuidv4(),
        amount: 100.50,
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Forwarded-For': myIp,
        },
    };

    const res = http.post(url, payload, params);

    if (res.status === 200) status200.add(1);
    else if (res.status === 409) status409.add(1);
    else if (res.status === 429) status429.add(1);
    else if (res.status >= 500) status500.add(1);

    sleep(0.1);
}