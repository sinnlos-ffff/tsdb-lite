import http from "k6/http";
import { check, fail } from "k6";
import exec from "k6/execution";

export const options = {
  scenarios: {
    main: {
      executor: 'per-vu-iterations',
      vus: 100,
      iterations: 100,
      maxDuration: '1h',
    },
  },
};

const url = "http://localhost:8080/point";

export function setup() {
  const setupUrl = "http://localhost:8080/timeseries";
  const seriesCount = 100;
  const timeSeries = [];

  for (let i = 0; i < seriesCount; i++) {
    const metric = `test_metric_${i}`;
    const tags = { host: `host_${i}` };
    const payload = JSON.stringify({
      metric: metric,
      tags: tags,
    });
    const params = {
      headers: { "Content-Type": "application/json" },
    };
    const res = http.post(setupUrl, payload, params);
    if (res.status !== 200) {
      fail(
        `Failed to create time series during setup. Status: ${res.status}, Body: ${res.body}`
      );
    }
    timeSeries.push({ metric, tags });
  }
  return { timeSeries };
}

export default function (data) {
  const seriesIndex = exec.scenario.iterationInTest % data.timeSeries.length;
  const series = data.timeSeries[seriesIndex];

  const payload = JSON.stringify({
    metric: series.metric,
    timestamp: Math.floor(Date.now() / 1000),
    value: Math.random() * 100,
    tags: series.tags,
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
}