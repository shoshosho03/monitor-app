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
    <div style={{ padding: 20 }}>
      <h1>System Monitor</h1>

      <p>CPU: {stats.cpu.toFixed(2)}%</p>
      <p>Memory: {stats.memory.toFixed(2)}%</p>

      {error && <p style={{ color: "red" }}>{error}</p>}

      <hr />

      <div>
        <label>更新間隔(ms): </label>
        <input
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
      </div>
    </div>
  );
}

export default App;
