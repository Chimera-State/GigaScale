import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    insecureSkipTLSVerify: true,
    stages: [
        { duration: '5s', target: 2000 },
        { duration: '10s', target: 10000 },
        { duration: '5s', target: 0 },
    ],
};

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

// Havuz mantığı 
function generateRandomIP(vuId) {
    const ipPoolIndex = vuId % 200;
    return `192.168.100.${ipPoolIndex}`;
}

const vusIPs = {};

export default function () {
    // Her VU için kalıcı IP atanması
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
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Forwarded-For': myIp,
        },
    };

    const res = http.post(url, payload, params);

    check(res, {
        'sistem 500 donmedi (200/409/429 kabul edilebilir)': (r) =>
            r.status === 200 || r.status === 409 || r.status === 429,
    });

    sleep(0.1);
}