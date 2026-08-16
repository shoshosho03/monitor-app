# monitor-app

CPU使用率とメモリ使用率を定期取得し、現在値と過去10分間の推移を表示するデスクトップモニターです。黒地に緑色の発光表示、7セグメント数字、折れ線グラフを組み合わせた小型ウィンドウとして動作します。

- デスクトップアプリ基盤: Wails v2
- バックエンド: Go / gopsutil
- フロントエンド: React 18 / Vite
- 主な対象環境: Windows x64

## 主な機能

- CPU全体の使用率をパーセント表示
- メモリ使用率をパーセント表示
- CPU・メモリそれぞれの過去10分間の履歴グラフ
- 取得間隔をミリ秒単位で変更
- CSSで作成した7セグメント数字
- OS標準タイトルバーを使わないフレームレス表示
- ヘッダーのドラッグによるウィンドウ移動
- Windowsのタスクバーにボタンを表示しないツールウィンドウ動作

## 処理の流れ

```text
main.go
  Wailsを起動してAppをフロントエンドへ公開
    ↓
app.go
  起動コンテキストを保存し、Windows用ウィンドウ設定を反映
    ↓
frontend/src/App.jsx
  初回表示時と指定間隔ごとにGetStats()を呼び出す
    ↓
frontend/wailsjs/go/main/App.js
  JavaScript呼び出しをGoのApp.GetStats()へ橋渡し
    ↓
app.go
  GetSystemStats()を呼び、{ cpu, memory }へ整形
    ↓
monitor.go
  gopsutilでOSからCPU・メモリ使用率を取得
    ↓
App.jsx
  現在値と直近10分の履歴を更新し、7セグメント数字とSVGグラフを再描画
```

Go側でエラーが発生すると、WailsによってJavaScriptのPromiseがrejectされ、ヘッダーに `取得エラー` が表示されます。

## ディレクトリ構成

```text
monitor-app/
├── main.go                     # Wailsの起動とウィンドウ設定
├── app.go                      # Wailsのライフサイクルと公開API
├── monitor.go                  # CPU・メモリ情報の取得
├── taskbar_windows.go          # Windows固有のウィンドウスタイル変更
├── taskbar_other.go            # Windows以外向けの代替実装
├── go.mod / go.sum             # Goモジュールと依存関係
├── wails.json                  # Wails CLIの設定
├── frontend/
│   ├── index.html              # Reactを読み込むHTML
│   ├── package.json            # npmスクリプトと依存関係
│   ├── vite.config.js          # Vite設定
│   ├── src/
│   │   ├── main.jsx            # Reactのエントリーポイント
│   │   ├── App.jsx             # 画面、状態管理、履歴グラフ
│   │   ├── style.css           # レイアウトとデジタル表示
│   │   └── App.css             # 未使用のテンプレート由来CSS
│   ├── wailsjs/                # Wailsが生成するGo/JSブリッジ
│   ├── dist/                   # Viteのビルド出力
│   └── node_modules/           # npm依存パッケージ
├── build/
│   ├── appicon.png             # アプリアイコンの元画像
│   ├── windows/                # Windows用アイコン、manifest、installer設定
│   └── bin/                    # 生成された実行ファイル
└── README.md
```

`frontend/dist`、`frontend/node_modules`、`build/bin` は生成物です。`frontend/wailsjs` もWailsが生成するため、通常は直接編集しません。

## Goバックエンド

### `main.go`

アプリのエントリーポイントです。

#### フロントエンドの埋め込み

```go
//go:embed all:frontend/dist
var assets embed.FS
```

Viteが生成した `frontend/dist` を実行ファイルへ埋め込みます。このため、Windowsへ配布するときはHTMLやJavaScriptを別ファイルで添付する必要がありません。

#### ウィンドウサイズ

```go
const (
    windowWidth  = 420
    windowHeight = 147
)
```

起動時とDOM準備完了後の再設定で同じ値を使うため、幅と高さを定数化しています。

#### Wailsオプション

- `Width` / `Height`: 初期ウィンドウサイズ
- `MinHeight: 1`: Windows標準の最小高より小さなサイズを指定できるようにする設定
- `Frameless: true`: OS標準タイトルバーを非表示にする
- `DisableResize: true`: ユーザーによるリサイズを無効化する
- `BackgroundColour`: WebView表示前にもテーマと同じ暗色を表示する
- `WindowClassName`: Win32 APIから対象ウィンドウを探すためのクラス名
- `DisableFramelessWindowDecorations`: Windowsの影や丸角などの標準装飾を無効化する
- `OnStartup`: `app.startup` を登録する
- `OnDomReady`: `app.domReady` を登録する
- `Bind`: `App` の公開メソッドをJavaScriptから呼べるようにする

`AssetServer` には埋め込んだファイルシステムを渡します。最後に `wails.Run` がネイティブウィンドウとWebViewを起動します。

### `app.go`

Wailsのライフサイクルとアプリ固有処理の境界です。

#### `App`

```go
type App struct {
    ctx context.Context
}
```

Wailsランタイムを操作するためのコンテキストを保持します。

#### `NewApp`

`App` をポインタで生成するコンストラクタです。Wailsの `Bind` にはこのインスタンスを渡します。

#### `startup`

Wails起動時に呼ばれ、受け取った `context.Context` を `App` に保存します。

#### `domReady`

WebViewのDOM準備完了後に呼ばれます。

1. `hideWindowFromTaskbar()` でWindows用ウィンドウスタイルを変更
2. 失敗時はログへ警告を出力
3. `runtime.WindowSetSize` で意図したサイズを再適用

サイズを再適用する理由は、Windowsのネイティブウィンドウスタイルを変更した後に、最終的なクライアント領域を指定値へ合わせるためです。

#### `GetStats`

Reactから呼び出される公開APIです。`GetSystemStats()` の戻り値を次のオブジェクトへ変換します。

```json
{
  "memory": 42.5,
  "cpu": 13.7
}
```

Goの `error` はWailsによってJavaScriptのPromiseエラーへ変換されます。

### `monitor.go`

画面やWailsに依存しない、システム情報取得ロジックです。

#### メモリ使用率

```go
vmStat, err := mem.VirtualMemory()
```

物理メモリ情報を取得し、`vmStat.UsedPercent` を使用率として返します。

#### CPU使用率

```go
cpuPercent, err := cpu.Percent(0, false)
```

- 第1引数 `0`: 関数内で待機せず、前回値との差分から使用率を計算
- 第2引数 `false`: CPUコア別ではなくシステム全体を1件で返す

結果はスライスなので、全体使用率として `cpuPercent[0]` を返します。戻り値の順番は「メモリ使用率、CPU使用率、エラー」です。

### `taskbar_windows.go`

`//go:build windows` が付いているため、Windows向けビルドでのみ使われます。標準ライブラリの `syscall` を通して `user32.dll` を呼び出します。

#### 対象ウィンドウの取得

`FindWindowW` と `windowClassName` を使い、Wailsが生成したネイティブウィンドウのハンドルを取得します。

#### 通常ウィンドウスタイルの解除

`GetWindowLongPtrW` で現在の `GWL_STYLE` を読み、`WS_OVERLAPPEDWINDOW` を外して `WS_POPUP | WS_VISIBLE` を設定します。

これにより、見た目だけでなくWin32上でもポップアップ形式のフレームレスウィンドウになります。

#### タスクバーボタンの非表示

拡張スタイル `GWL_EXSTYLE` に対して次の変更を行います。

- `WS_EX_TOOLWINDOW` を追加
- `WS_EX_APPWINDOW` を削除

この組み合わせで、ウィンドウを表示したままタスクバーのボタンを非表示にします。

#### スタイル変更の反映

`SetWindowPos` をサイズ・位置変更なしで呼び、`SWP_FRAMECHANGED` を指定します。これにより、変更したウィンドウスタイルをWindowsへ再評価させます。

各Win32 API呼び出しは失敗を確認し、呼び出し元へ文脈付きのエラーを返します。

### `taskbar_other.go`

`//go:build !windows` が付いているため、Windows以外で使われます。

```go
func hideWindowFromTaskbar() error {
    return nil
}
```

Windows固有の関数と同じシグネチャを持つ何もしない実装です。これにより、`app.go` にOS判定を書かず、LinuxやmacOSでもコンパイルできます。

### `go.mod` / `go.sum`

独立したGoモジュールです。`go.mod` ではGo 1.24.0を指定しています。

主な直接依存は次の2つです。

- `github.com/wailsapp/wails/v2`: デスクトップアプリ基盤
- `github.com/shirou/gopsutil/v4`: CPU・メモリ情報の取得

`go.sum` は依存モジュールのチェックサムを管理します。

## Reactフロントエンド

### `frontend/index.html`

WebViewへ読み込まれるHTMLの土台です。Reactの描画先となる `#root` を用意し、`src/main.jsx` をES Modulesとして読み込みます。

### `frontend/src/main.jsx`

React 18の `createRoot` で `App` を描画し、全体スタイルの `style.css` を読み込みます。

開発時は `React.StrictMode` が有効です。副作用を検査するため、開発環境では `useEffect` の実行、クリーンアップ、再実行が行われ、初回API呼び出しが複数回に見える場合があります。

### `frontend/src/App.jsx`

画面、状態管理、定期取得、7セグメント表示、履歴グラフを実装しています。

#### `digitSegments`

各数字を構成する7本のセグメントを定義した対応表です。

```text
  a
f   b
  g
e   c
  d
```

例えば `1` は `b` と `c`、`8` は `a`〜`g` の全セグメントを点灯します。

#### `DigitalNumber`

数値文字列を1文字ずつ分解し、各桁に7本の `<i>` 要素を生成します。点灯対象には `is-on` クラスを付け、CSSで緑色に発光させます。

- `value`: 表示する数値または入力中の文字列
- `fractionDigits`: 小数点以下の桁数。CPU・MEMは既定値の1桁
- `small`: SPEED欄用の小型表示

小数点は専用の `digital-number__dot` で描画します。見た目のセグメントは `aria-hidden` にし、外側へ `aria-label` を付けることで読み上げ可能な数値を保持しています。

#### `historyWindowMs`

```js
const historyWindowMs = 10 * 60 * 1000;
```

履歴として保持する10分間をミリ秒で表します。

#### `UsageGraph`

履歴データをSVG折れ線グラフへ変換します。

横座標は、10分前を `0`、現在を `100` とする次の比率で計算します。

```text
x = (サンプル時刻 - 10分前の時刻) / 10分 × 100
```

縦座標は使用率を `0`〜`100%` に制限し、SVGの上端が100%、下端が0%になるよう反転します。

```text
y = 28 - 使用率 / 100 × 28
```

- `<line>`: 25%、50%、75%の補助線
- `<polyline>`: 使用率の履歴
- `<circle>`: 最新値の位置
- `−10M` / `NOW`: 時間軸の両端

`preserveAspectRatio="none"` により、カード幅に合わせてグラフを横方向へ伸縮します。

#### React state

- `stats`: 最新のCPU・メモリ使用率
- `history`: `{ time, cpu, memory }` 形式の履歴
- `intervalMs`: 実際にタイマーへ適用する更新間隔
- `intervalInput`: SPEED欄へ表示する入力中の文字列
- `error`: 取得失敗時のメッセージ

`intervalMs` と `intervalInput` を分けているため、Backspaceで入力欄を空にしても直前の有効なタイマー値は維持されます。

#### 定期取得

`useEffect` 内の `fetchStats` が次の処理を行います。

1. `GetStats()` でGo側から最新値を取得
2. `stats` を更新
3. 現在時刻より10分以上前の履歴を削除
4. 最新値を時刻付きで履歴へ追加
5. エラー表示を解除

初回は即時実行し、その後 `setInterval` で繰り返します。更新間隔が変わると古いタイマーを `clearInterval` で破棄し、新しい間隔で作り直します。

履歴はメモリ上だけに保持するため、アプリを終了すると消去されます。

#### SPEED入力

HTML上の設定は最小 `500ms`、刻み幅 `500ms` です。入力中の文字列は常に `intervalInput` へ反映されるため、Backspaceで空欄にできます。

JavaScript側では `100ms` 以上の値だけを `intervalMs` へ反映します。フォーカスを外した時点で100ms未満なら、最後に有効だった値へ戻します。

透明な `<input type="number">` を7セグメント表示へ重ねているため、入力機能を保ちながら見た目をデジタル表示にしています。

#### ヘッダーと終了

- `RESOURCE MONITOR`: アプリ名
- `LIVE`: 正常取得中の状態
- `取得エラー`: Go APIの呼び出し失敗時
- `×`: Wailsランタイムの `Quit()` を呼んでアプリを終了

### `frontend/src/style.css`

画面全体のデザインと7セグメント表示を定義します。

#### 全体テーマ

- 背景: `#020604`
- 前景: `#58f56d`
- OCR系・等幅フォントを優先
- 半透明グラデーションで走査線風の背景を作成
- `text-shadow` と `box-shadow` で発光を表現

#### フレームレスウィンドウの操作

```css
.monitor__header {
    --wails-draggable: drag;
}

.monitor__close {
    --wails-draggable: no-drag;
}
```

ヘッダーをドラッグ領域にし、終了ボタンだけをドラッグ対象外にします。

#### 7セグメント

各セグメントを絶対配置し、横棒・縦棒を `clip-path` で六角形に近い形へ加工しています。消灯中も薄く表示し、`is-on` のセグメントだけを強く発光させます。

`digital-number--small` はSPEED欄用に幅と高さを縮めたバリエーションです。

#### 履歴グラフ

`stat__graph` がグラフの枠、SVGが描画領域です。補助線は薄く、折れ線と最新点は明るく表示します。`vector-effect: non-scaling-stroke` により、SVGが伸縮しても線幅が極端に変化しません。

### `frontend/wailsjs/`

Wailsが生成するJavaScript・TypeScriptコードです。

- `go/main/App.js`: `window.go.main.App.GetStats` を呼ぶラッパー
- `go/main/App.d.ts`: 公開APIの型定義
- `runtime/`: `Quit` などWailsランタイムAPIのラッパー

Go側の公開メソッドを変更した場合は、`wails dev` または `wails build` で再生成します。

### `frontend/dist/`

`npm run build` が生成する本番用HTML、JavaScript、CSSです。`main.go` の `go:embed` 対象になります。

## 設定ファイル

### `wails.json`

Wails CLIが参照する設定です。

- `name`: Wailsプロジェクト名
- `outputfilename`: 実行ファイル名
- `frontend:install`: フロントエンド依存関係のインストールコマンド
- `frontend:build`: 本番フロントエンドのビルドコマンド
- `frontend:dev:watcher`: 開発サーバーの起動コマンド
- `frontend:dev:serverUrl`: Wailsが開発サーバーを自動検出

### `frontend/package.json`

- `npm run dev`: Vite開発サーバー
- `npm run build`: 本番用ファイルを `dist` へ生成
- `npm run preview`: ビルド結果のプレビュー

ReactとReact DOMが実行時依存、ViteとReactプラグインが開発時依存です。

### `build/windows/`

- `icon.ico`: Windows実行ファイルのアイコン
- `info.json`: バージョン情報リソースのテンプレート
- `wails.exe.manifest`: DPIやWindows互換性などのmanifest
- `installer/`: NSISインストーラー用スクリプト

## 開発と検証

コマンドはこのディレクトリで実行します。

```bash
cd monitor/monitor-app
```

### 開発モード

```bash
wails dev
```

GoバックエンドとVite開発サーバーを起動します。フロントエンドの変更はホットリロードされます。ネイティブウィンドウ設定の変更は、アプリの完全な再起動が必要になる場合があります。

### フロントエンドのみビルド

```bash
cd frontend
npm install
npm run build
```

成果物は `frontend/dist` に生成されます。

### Goの検証

```bash
go test ./...
go vet ./...
```

現時点ではテストファイルがないため、`go test` は主に全Goパッケージがコンパイルできることを確認します。

Windows向けコードだけをLinux上でコンパイル確認する場合は次を使用できます。

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test ./...
```

## Windows実行ファイルの作成

### Windows上でビルド

```bash
wails build -platform windows/amd64 -clean
```

### Linux上でクロスビルド

```bash
GOCACHE=/tmp/monitor-app-go-cache \
  wails build -platform windows/amd64 -clean
```

生成先は次のとおりです。

```text
build/bin/monitor-app.exe
```

生成された `.exe` はPE32+形式のWindows x64 GUIアプリです。Windowsでの表示にはMicrosoft Edge WebView2 Runtimeが必要です。一般的なWindows 10・11環境では導入済みですが、未導入の場合は別途インストールが必要です。

実行ファイルにはコード署名を行っていないため、配布先の設定によってはSmartScreen警告が表示されます。
