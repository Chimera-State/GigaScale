import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,
    duration: '30s',


    thresholds: {
        http_req_failed: ['rate<0.01'],
        http_req_duration: ['p(95)<200'],
    },
};

export default function () {

    const res = http.get('http://gateway:8080/health');

    check(res, {
        'status 200 (Gateway Ayakta)': (r) => r.status === 200,
        'cevap süresi 500ms altında': (r) => r.timings.duration < 500,
    });

    sleep(1);
}
