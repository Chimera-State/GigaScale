import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10, // virtual users
    duration: '30s', // testin süresi
};

export default function () {
    const res =
        http.get('http://gateway:8080/api/v1/reserve');
    check(res, {
        'status 200 mü?': (r) => r.status === 200,
    });
    sleep(1);
}