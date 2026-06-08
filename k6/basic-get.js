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
        'Status_200_Success': (r) => r.status === 200,
        'Response_Time_Under_500ms': (r) => r.timings.duration < 500,
    });

    sleep(1);
}
