package main

import (
	"context"
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
