import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  scenarios: {
    wrk_like: {
      executor: "constant-vus",
      vus: 300,
      duration: "30s",
    },
  },

  // penting untuk tail latency
  summaryTrendStats: [
    "avg",
    "min",
    "med",
    "p(75)",
    "p(90)",
    "p(95)",
    "p(99)",
    "max",
  ],

  // jangan limit request rate (biar sebrutal wrk)
  noConnectionReuse: false,
  insecureSkipTLSVerify: true,
};

// The default exported function is gonna be picked up by k6 as the entry point for the test script. It will be executed repeatedly in "iterations" for the whole duration of the test.
export default function () {
  // Make a GET request to the target URL
  const res = http.get("http://testing.local/health");

  check(res, { "status was 200": (r) => r.status == 200 });
}
