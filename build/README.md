# ビルドディレクトリ

このディレクトリには、アプリケーションのビルドに使うファイルと素材を配置します。

構成は次のとおりです。

- `bin`: ビルドした実行ファイルの出力先
- `darwin`: macOS 固有のファイル
- `windows`: Windows 固有のファイル

## macOS

`darwin` ディレクトリには、macOS 向けビルドで使うファイルがあります。アプリケーションに合わせて変更できます。初期状態へ戻す場合は対象ファイルを削除し、`wails build` を実行すると再生成されます。

- `Info.plist`: `wails build` で使うアプリケーション情報ファイル
- `Info.dev.plist`: `wails dev` で使う開発用のアプリケーション情報ファイル

## Windows

`windows` ディレクトリには、`wails build` が参照するマニフェストやリソース関連ファイルがあります。アプリケーションに合わせて変更できます。初期状態へ戻す場合は対象ファイルを削除し、`wails build` を実行すると再生成されます。

- `icon.ico`: Windows 版のアプリアイコン。存在しない場合は `build/appicon.png` から生成されます
- `installer/*`: Windows インストーラーの作成に使うファイル
- `info.json`: 実行ファイルのプロパティやインストーラーに反映するアプリケーション情報
- `wails.exe.manifest`: Windows 版のアプリケーションマニフェスト
