//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	windowClassName = "monitorAppWindow"
	gwlStyle        = -16
	gwlExStyle      = -20
	wsOverlapped    = 0x00CF0000
	wsPopup         = 0x80000000
	wsVisible       = 0x10000000
	wsExToolWindow  = 0x00000080
	wsExAppWindow   = 0x00040000
	swpNoSize       = 0x0001
	swpNoMove       = 0x0002
	swpNoZOrder     = 0x0004
	swpNoActivate   = 0x0010
	swpFrameChanged = 0x0020
	spiGetWorkArea  = 0x0030
	windowMargin    = 8
)

type rect struct {
	left   int32
	top    int32
	right  int32
	bottom int32
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	findWindowW       = user32.NewProc("FindWindowW")
	getWindowLongPtrW = user32.NewProc("GetWindowLongPtrW")
	setWindowLongPtrW = user32.NewProc("SetWindowLongPtrW")
	setWindowPos      = user32.NewProc("SetWindowPos")
	getWindowRect     = user32.NewProc("GetWindowRect")
	systemParameters  = user32.NewProc("SystemParametersInfoW")
)

func findAppWindow() (uintptr, error) {
	className, err := syscall.UTF16PtrFromString(windowClassName)
	if err != nil {
		return 0, fmt.Errorf("ウィンドウクラス名の変換: %w", err)
	}

	hwnd, _, _ := findWindowW.Call(uintptr(unsafe.Pointer(className)), 0)
	if hwnd == 0 {
		return 0, fmt.Errorf("対象ウィンドウが見つかりません")
	}

	return hwnd, nil
}

func hideWindowFromTaskbar() error {
	hwnd, err := findAppWindow()
	if err != nil {
		return err
	}

	// フレームレス表示でも残る通常ウィンドウスタイルを外し、
	// Windows標準の最小ウィンドウ高が適用されないようにする。
	styleIndex := int32(gwlStyle)
	style, _, callErr := getWindowLongPtrW.Call(hwnd, uintptr(styleIndex))
	if style == 0 && callErr != syscall.Errno(0) {
		return fmt.Errorf("ウィンドウスタイルの取得: %w", callErr)
	}

	newWindowStyle := (style &^ uintptr(wsOverlapped)) | uintptr(wsPopup|wsVisible)
	previousStyle, _, callErr := setWindowLongPtrW.Call(hwnd, uintptr(styleIndex), newWindowStyle)
	if previousStyle == 0 && callErr != syscall.Errno(0) {
		return fmt.Errorf("ウィンドウスタイルの更新: %w", callErr)
	}

	index := int32(gwlExStyle)
	exStyle, _, callErr := getWindowLongPtrW.Call(hwnd, uintptr(index))
	if exStyle == 0 && callErr != syscall.Errno(0) {
		return fmt.Errorf("拡張スタイルの取得: %w", callErr)
	}

	// WS_EX_TOOLWINDOW を付け、WS_EX_APPWINDOW を外すとタスクバーに表示されない。
	newExStyle := (exStyle | wsExToolWindow) &^ wsExAppWindow
	previousExStyle, _, callErr := setWindowLongPtrW.Call(hwnd, uintptr(index), newExStyle)
	if previousExStyle == 0 && callErr != syscall.Errno(0) {
		return fmt.Errorf("拡張スタイルの更新: %w", callErr)
	}

	flags := uintptr(swpNoSize | swpNoMove | swpNoZOrder | swpNoActivate | swpFrameChanged)
	result, _, callErr := setWindowPos.Call(hwnd, 0, 0, 0, 0, 0, flags)
	if result == 0 {
		return fmt.Errorf("ウィンドウスタイルの反映: %w", callErr)
	}

	return nil
}

func moveWindowToTopRight() error {
	hwnd, err := findAppWindow()
	if err != nil {
		return err
	}

	var workArea rect
	result, _, callErr := systemParameters.Call(
		spiGetWorkArea,
		0,
		uintptr(unsafe.Pointer(&workArea)),
		0,
	)
	if result == 0 {
		return fmt.Errorf("画面の作業領域の取得: %w", callErr)
	}

	var windowRect rect
	result, _, callErr = getWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&windowRect)))
	if result == 0 {
		return fmt.Errorf("ウィンドウサイズの取得: %w", callErr)
	}

	width := windowRect.right - windowRect.left
	x := workArea.right - width - windowMargin
	y := workArea.top + windowMargin
	flags := uintptr(swpNoSize | swpNoZOrder | swpNoActivate)
	result, _, callErr = setWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), 0, 0, flags)
	if result == 0 {
		return fmt.Errorf("ウィンドウ位置の更新: %w", callErr)
	}

	return nil
}
