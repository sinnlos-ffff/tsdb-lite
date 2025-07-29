import http from "k6/http";
import { check, sleep, fail } from "k6";

export const options = {
  vus: 1000,
  duration: "10s",
};

const url = "http://localhost:8080/point";

export function setup() {
  const setupUrl = "http://localhost:8080/timeseries";
  const payload = JSON.stringify({
    metric: "test_metric",
    tags: { tag1: "value1" },
  });
  const params = {
    headers: { "Content-Type": "application/json" },
  };
  const res = http.post(setupUrl, payload, params);
  if (res.status !== 200) {
    fail(
      `Failed to create time series during setup. Status: ${res.status}, Body: ${res.body}`,
    );
  }
}

export default function () {
  const payload = JSON.stringify({
    metric: "test_metric",
    timestamp: Math.floor(Date.now() / 1000),
    value: Math.random() * 100,
    tags: { tag1: "value1" },
  });

  const params = {
    headers: { "Content-Type": "application/json" },
  };

  const res = http.post(url, payload, params);
  const ok = check(res, {
    "status was 200": (r) => r.status === 200,
  });

  if (!ok) {
    console.error(`Failed request: ${res.status}, Body: ${res.body}`);
  }

  sleep(1);
}
