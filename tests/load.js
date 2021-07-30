import http from 'k6/http';
import { sleep } from 'k6';
import { randomIntBetween } from "https://jslib.k6.io/k6-utils/1.1.0/index.js";

export let options = {
   vus: 100,
   duration: '30s'
}
	
export default function () {
   for (var i=0; i<100; i++ ) {
      let Die = randomIntBetween(1,100);
      switch (true) {
         case (Die <= 60): 
            http.get('http://127.0.0.1/fib/'+randomIntBetween(1,80));
            break;
         case (Die > 60 && Die <= 99): 
            http.get('http://127.0.0.1/fetchmemoct//'+randomIntBetween(1,999999));
            break;
         case Die > 99: 
            http.get('http://127.0.0.1/purge');
            break;
      }
   }
}
