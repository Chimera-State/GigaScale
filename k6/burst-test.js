import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
    insecureSkipTLSVerify: true,
    stages: [
        { duration: '10s', target: 500 },
    ],
};

const generateAlphanumID = () => uuidv4().replace(/-/g, '');

function getRandomIP() {
    return `${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}.${Math.floor(Math.random() * 255)}`;
}

export default function () {
    const url = 'https://api.yusufyonturk.com/api/v1/reserve'; 

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
            'X-Forwarded-For': getRandomIP(),
        },
    };

    const res = http.post(url, payload, params);

    check(res, { 
        'status 200': (r) => r.status === 200,  
        'status 409': (r) => r.status === 409,
        'status 429': (r) => r.status === 429,
        'status 500': (r) => r.status === 500,
    });
}
