# monitor-app

CPU 使用率とメモリ使用率を一定間隔で取得し、デスクトップウィンドウに表示する学習用アプリです。

- デスクトップアプリ基盤: Wails v2
- バックエンド: Go / gopsutil
- フロントエンド: React 18 / Vite

Wails のバインディングを介して React から Go のメソッドを呼び出し、OS から取得したシステム情報を画面へ返します。

## 処理の流れ

```text
main.go
  Wails を起動し、App をフロントエンドへ公開
    ↓
frontend/src/App.jsx
  初回表示時と指定間隔ごとに GetStats() を呼び出す
    ↓
frontend/wailsjs/go/main/App.js
  JavaScript の呼び出しを Go の App.GetStats() へ橋渡し
    ↓
app.go
  GetSystemStats() の結果を { cpu, memory } の形に整える
    ↓
monitor.go
  gopsutil で CPU・メモリ使用率を取得
    ↓
App.jsx
  React の state を更新して最新値を再描画
```

Go 側でエラーが発生すると JavaScript 側の Promise が reject され、画面に `取得エラー` と表示されます。

## ディレクトリ構成

```text
monitor-app/
├── main.go                     # Wails アプリのエントリーポイント
├── app.go                      # React に公開する API とアプリの状態
├── monitor.go                  # CPU・メモリ情報を取得するロジック
├── go.mod / go.sum             # Go のバージョンと依存モジュール
├── wails.json                  # Wails のプロジェクト・ビルド設定
├── frontend/
│   ├── index.html              # React を読み込む HTML
│   ├── package.json            # npm スクリプトとフロントエンド依存関係
│   ├── package-lock.json       # npm 依存関係の固定情報
│   ├── package.json.md5        # Wails が依存関係の変更検知に使う値
│   ├── vite.config.js          # Vite と React プラグインの設定
│   ├── src/
│   │   ├── main.jsx            # React の起動処理
│   │   ├── App.jsx             # 表示、状態管理、定期取得処理
│   │   ├── style.css           # 現在読み込まれる全体スタイル
│   │   ├── App.css             # テンプレート由来のスタイル（現在は未使用）
│   │   └── assets/             # フォント、画像、ライセンス
│   ├── wailsjs/                # Wails が生成する Go/JS ブリッジ
│   ├── dist/                   # Vite のビルド出力
│   └── node_modules/           # npm がインストールする依存パッケージ
├── build/
│   ├── appicon.png             # アプリの元アイコン
│   ├── darwin/                 # macOS 用 plist
│   ├── windows/                # Windows 用アイコン、manifest、installer 設定
│   └── bin/                    # ビルドした実行ファイルの出力先
├── .gitignore                  # 生成物などの Git 除外設定
└── README.md                   # このドキュメント
```

`frontend/dist`、`frontend/node_modules`、`build/bin` は生成物であり、`.gitignore` の対象です。`frontend/wailsjs` も Wails により生成されるため、通常は直接編集しません。

## Go バックエンド

### `main.go`

アプリの開始地点です。主に次の処理を担当します。

- `//go:embed all:frontend/dist` で Vite のビルド結果を実行ファイルへ埋め込む
- `NewApp()` で `App` を生成する
- ウィンドウタイトル、サイズ、背景色を設定する
- `OnStartup` に `app.startup` を登録する
- `Bind` に `app` を登録し、公開メソッドを JavaScript から呼べるようにする
- `wails.Run` でアプリを起動する

本番ビルドでは埋め込まれた `frontend/dist` が WebView に配信されます。

### `app.go`

Wails とアプリ固有ロジックの境界です。

- `App`: Wails の実行コンテキストを保持する構造体
- `NewApp`: `App` を生成するコンストラクタ
- `startup`: Wails の起動時に受け取った `context.Context` を保存するコールバック
- `GetStats`: React から呼び出せる公開 API

`GetStats` は `monitor.go` の `GetSystemStats` を呼び出し、成功時に次の形で値を返します。

```json
{
  "memory": 42.5,
  "cpu": 13.7
}
```

Go の `error` は Wails によって JavaScript の Promise のエラーへ変換されます。

### `monitor.go`

画面や Wails に依存しないシステム情報取得ロジックです。`GetSystemStats` は次の順で値を返します。

1. メモリ使用率（%）
2. CPU 使用率（%）
3. エラー

メモリは `mem.VirtualMemory()` の `UsedPercent`、CPU は `cpu.Percent(0, false)` で取得します。CPU はコア別ではなく全体の使用率です。いずれかの取得に失敗した場合は、そのエラーを呼び出し元へ返します。

### `go.mod` / `go.sum`

このアプリは独立した Go モジュールです。`go.mod` では Go 1.24.0 を指定し、主に次の直接依存を管理しています。

- `github.com/wailsapp/wails/v2`: デスクトップアプリ基盤
- `github.com/shirou/gopsutil/v4`: CPU・メモリ情報の取得

`go.sum` は依存モジュールのチェックサムを記録します。

### `wails.json`

Wails CLI が参照するプロジェクト設定です。アプリ名、出力ファイル名、フロントエンドのインストール・ビルド・開発コマンドなどを定義します。

## React フロントエンド

### `frontend/index.html`

WebView に表示する HTML の土台です。`#root` 要素を用意し、`src/main.jsx` を ES Modules として読み込みます。

### `frontend/src/main.jsx`

React 18 の `createRoot` を使い、`#root` に `App` コンポーネントを描画します。開発時の問題を検出しやすくするため `React.StrictMode` を有効にしています。

Strict Mode の開発環境では、副作用の確認のため `useEffect` が実行、クリーンアップ、再実行され、初回の API 呼び出しが複数回に見えることがあります。

### `frontend/src/App.jsx`

画面と定期更新処理の中心です。次の state を管理します。

- `stats`: CPU・メモリ使用率。初期値はどちらも `0`
- `history`: 取得時刻付きのCPU・メモリ使用率。グラフ表示用に直近10分だけ保持
- `intervalMs`: 更新間隔。初期値は `1000` ミリ秒
- `error`: 取得失敗時に表示するメッセージ

`useEffect` は初回表示時と更新間隔の変更時に動きます。まず `GetStats()` を即時実行し、その後は `setInterval` で繰り返します。effect のクリーンアップでは `clearInterval` を呼ぶため、間隔変更やコンポーネント破棄後に古いタイマーが残りません。

CPU・メモリの取得結果は時刻とともに保存され、10分を過ぎたデータを削除しながらSVGの折れ線グラフとして表示します。縦軸は使用率 `0`〜`100%`、横軸は左端が10分前、右端が現在です。

入力欄は数値をミリ秒として受け取り、表示上は最小値 `500`、刻み幅 `500` です。イベント処理では空値または `100` 未満の値を state に反映しません。

### `frontend/src/style.css` / `App.css`

`style.css` は `main.jsx` から読み込まれ、背景色、文字色、フォントなどの全体スタイルを定義します。`App.css` は Wails テンプレート由来ですが、現在のコードからは import されていないため表示には適用されません。

### `frontend/wailsjs/`

Wails が生成する JavaScript/TypeScript コードです。

- `go/main/App.js`: `window.go.main.App.GetStats` を呼ぶ JavaScript ラッパー
- `go/main/App.d.ts`: 公開メソッドの TypeScript 型定義
- `runtime/`: Wails ランタイム API と型定義

Go 側の公開メソッドを変更した場合は、Wails の開発・ビルド処理を通して再生成します。

### `frontend/dist/`

`npm run build` が生成する HTML、JavaScript、CSS です。`main.go` の `go:embed` 対象であり、ソースファイルではありません。

## build ディレクトリ

配布用バイナリや OS 固有設定を管理します。

- `appicon.png`: 各 OS 向けアイコンの生成元
- `darwin/Info.plist`: macOS の本番ビルド設定
- `darwin/Info.dev.plist`: macOS の開発実行設定
- `windows/icon.ico`: Windows のアプリアイコン
- `windows/info.json`: Windows のアプリ情報
- `windows/wails.exe.manifest`: Windows 実行ファイルの manifest
- `windows/installer/`: NSIS インストーラー設定
- `bin/`: `wails build` の出力先

詳しい OS 別ビルドファイルの説明は `build/README.md` にあります。

## 開発とビルド

コマンドはこのディレクトリで実行します。

```bash
cd monitor/monitor-app
```

### 開発モード

```bash
wails dev
```

Vite の開発サーバーと Wails アプリが起動し、フロントエンドの変更がホットリロードされます。

### フロントエンドのみビルド

```bash
cd frontend
npm install
npm run build
```

成果物は `frontend/dist` に作成されます。

### デスクトップアプリをビルド

```bash
wails build
```

`wails.json` の設定に従ってフロントエンドをビルドし、実行ファイルを `build/bin` に出力します。

### Go の検証

```bash
go test ./...
go vet ./...
```

現時点ではテストファイルがないため、`go test` は主に全 Go パッケージがコンパイルできることを確認します。
