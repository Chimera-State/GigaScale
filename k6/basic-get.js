import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10, // virtual users
    duration: '30s', // testin süresi
};

export default function () {
    const res = http.get('http://gateway:8080/health');
    check(res, {
        'gateway ayakta mı?': (r) => r.status === 200,
    });
    sleep(1);
}