import { useEffect, useState } from "react";
import { GetStats } from "../wailsjs/go/main/App";

function App() {
  const [stats, setStats] = useState({ cpu: 0, memory: 0 });
  const [intervalMs, setIntervalMs] = useState(1000);
  const [error, setError] = useState("");

  useEffect(() => {
    let timerId;

    const fetchStats = async () => {
      try {
        const result = await GetStats();
        setStats(result);
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
        <h1>System Monitor</h1>
        {error && <span className="monitor__error">{error}</span>}
      </div>

      <div className="monitor__content">
        <div className="monitor__stats">
          <div className="stat">
            <span className="stat__label">CPU</span>
            <strong>{stats.cpu.toFixed(1)}%</strong>
          </div>
          <div className="stat">
            <span className="stat__label">MEM</span>
            <strong>{stats.memory.toFixed(1)}%</strong>
          </div>
        </div>

        <label className="interval">
          <span>更新</span>
          <input
            aria-label="更新間隔（ミリ秒）"
            type="number"
            min="500"
            step="500"
            value={intervalMs}
            onChange={(e) => {
              const v = Number(e.target.value);
              if (!v || v < 100) return;
              setIntervalMs(v);
            }}
          />
          <span>ms</span>
        </label>
      </div>
    </main>
  );
}

export default App;
