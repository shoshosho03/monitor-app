//go:build !windows

package main

func hideWindowFromTaskbar() error {
	return nil
}

func moveWindowToTopRight() error {
	return nil
}
