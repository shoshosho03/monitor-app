package main

import (
	"context"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

// コンストラクタ（ポインタで返す）
func NewApp() *App {
	return &App{}
}

// 起動時にcontext受け取る
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// domReady hides the Windows taskbar button after the native window is ready.
func (a *App) domReady(ctx context.Context) {
	if err := hideWindowFromTaskbar(); err != nil {
		log.Printf("タスクバーアイコンを非表示にできませんでした: %v", err)
	}

	// Windowsの通常ウィンドウスタイルを外した後に、意図したサイズへ合わせる。
	runtime.WindowSetSize(ctx, windowWidth, windowHeight)

	if err := moveWindowToTopRight(); err != nil {
		log.Printf("ウィンドウを画面右上へ移動できませんでした: %v", err)
	}
}

// フロントから呼ばれるAPI
func (a *App) GetStats() (map[string]float64, error) {

	mem, cpu, err := GetSystemStats()
	if err != nil {
		return nil, err
	}

	return map[string]float64{
		"memory": mem,
		"cpu":    cpu,
	}, nil
}
