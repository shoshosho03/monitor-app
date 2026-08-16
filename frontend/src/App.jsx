import { useEffect, useState } from "react";
import { GetStats } from "../wailsjs/go/main/App";
import { Quit } from "../wailsjs/runtime/runtime";

const digitSegments = {
  0: "abcdef",
  1: "bc",
  2: "abdeg",
  3: "abcdg",
  4: "bcfg",
  5: "acdfg",
  6: "acdefg",
  7: "abc",
  8: "abcdefg",
  9: "abcdfg",
};

const historyWindowMs = 10 * 60 * 1000;

function DigitalNumber({ value, fractionDigits = 1, small = false }) {
  const text = typeof value === "string" ? value : Number(value).toFixed(fractionDigits);

  return (
    <span
      className={`digital-number${small ? " digital-number--small" : ""}`}
      aria-label={text}
    >
      {Array.from(text, (character, index) => character === "." ? (
        <span className="digital-number__dot" aria-hidden="true" key={index} />
      ) : (
        <span className="digital-number__digit" aria-hidden="true" key={index}>
          {Array.from("abcdefg", (segment) => (
            <i
              className={(digitSegments[character] || "").includes(segment) ? "is-on" : ""}
              data-segment={segment}
              key={segment}
            />
          ))}
        </span>
      ))}
    </span>
  );
}

function UsageGraph({ samples, metric, label }) {
  const now = Date.now();
  const start = now - historyWindowMs;
  const graphPoints = samples.map((sample) => {
    const x = Math.max(0, ((sample.time - start) / historyWindowMs) * 100);
    const y = 28 - (Math.min(100, Math.max(0, sample[metric])) / 100) * 28;
    return `${x.toFixed(2)},${y.toFixed(2)}`;
  });
  const points = graphPoints.join(" ");
  const latest = graphPoints.at(-1)?.split(",");

  return (
    <div className="stat__graph">
      <svg viewBox="0 0 100 28" preserveAspectRatio="none" role="img" aria-label={`${label}の過去10分間の使用率`}>
        <line x1="0" y1="7" x2="100" y2="7" />
        <line x1="0" y1="14" x2="100" y2="14" />
        <line x1="0" y1="21" x2="100" y2="21" />
        {points && <polyline points={points} />}
        {latest && <circle className="stat__graph-current" cx={latest[0]} cy={latest[1]} r="1.4" />}
      </svg>
      <span>−10M</span>
      <span>NOW</span>
    </div>
  );
}

function App() {
  const [stats, setStats] = useState({ cpu: 0, memory: 0 });
  const [history, setHistory] = useState([]);
  const [intervalMs, setIntervalMs] = useState(1000);
  const [intervalInput, setIntervalInput] = useState("1000");
  const [error, setError] = useState("");

  useEffect(() => {
    let timerId;

    const fetchStats = async () => {
      try {
        const result = await GetStats();
        setStats(result);
        setHistory((current) => {
          const now = Date.now();
          return [
            ...current.filter((sample) => sample.time >= now - historyWindowMs),
            { time: now, cpu: result.cpu, memory: result.memory },
          ];
        });
        setError("");
      } catch (e) {
        console.error(e);
        setError("取得エラー");
      }
    };

    // 初回実行
    fetchStats();

    // interval開始
    timerId = setInterval(fetchStats, intervalMs);

    // cleanup（超重要）
    return () => clearInterval(timerId);
  }, [intervalMs]);

  return (
    <main className="monitor">
      <div className="monitor__header">
        <h1><span aria-hidden="true">◆</span> RESOURCE MONITOR</h1>
        <div className="monitor__header-actions">
          {error ? (
            <span className="monitor__error">! {error}</span>
          ) : (
            <span className="monitor__status"><i aria-hidden="true" /> LIVE</span>
          )}
          <button className="monitor__close" type="button" onClick={Quit} aria-label="終了">
            ×
          </button>
        </div>
      </div>

      <div className="monitor__content">
        <div className="monitor__stats">
          <div className="stat">
            <div className="stat__row">
              <span className="stat__label">CPU</span>
              <strong><DigitalNumber value={stats.cpu} /><small>%</small></strong>
            </div>
            <UsageGraph samples={history} metric="cpu" label="CPU" />
          </div>
          <div className="stat">
            <div className="stat__row">
              <span className="stat__label">MEM</span>
              <strong><DigitalNumber value={stats.memory} /><small>%</small></strong>
            </div>
            <UsageGraph samples={history} metric="memory" label="メモリ" />
          </div>
        </div>

        <label className="interval">
          <span className="interval__label">SPEED</span>
          <span className="interval__input">
            <DigitalNumber value={intervalInput} fractionDigits={0} small />
            <input
              aria-label="更新間隔（ミリ秒）"
              type="number"
              min="500"
              step="500"
              value={intervalInput}
              onChange={(e) => {
                const nextValue = e.target.value;
                const interval = Number(nextValue);

                setIntervalInput(nextValue);
                if (interval >= 100) setIntervalMs(interval);
              }}
              onBlur={() => {
                if (Number(intervalInput) < 100) setIntervalInput(String(intervalMs));
              }}
            />
          </span>
          <span className="interval__unit">MS</span>
        </label>
      </div>
    </main>
  );
}

export default App;
